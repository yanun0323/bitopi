package service

import (
	"bitopi/internal/model"
	"context"
	"fmt"
	"net/http"
	"sort"

	"github.com/labstack/echo/v4"
)

const (
	_botServicePathKey = "service"
)

var (
	_validCategory = map[string]bool{
		"maid":   true,
		"pm":     true,
		"rails":  true,
		"devops": true,
	}
)

func DataResponse(c echo.Context, data interface{}, msgs ...string) error {
	msg := "OK"
	if len(msgs) != 0 {
		msg = msgs[0]
	}
	return c.JSON(http.StatusOK, struct {
		Msg  string      `json:"message"`
		Data interface{} `json:"data,omitempty"`
	}{
		Msg:  msg,
		Data: data,
	})
}

func ErrorResponse(c echo.Context, code int, msg string, errs ...error) error {
	errMsg := ""
	if len(errs) != 0 {
		errMsg = errs[0].Error()
	}

	return c.JSON(code, struct {
		Msg   string `json:"message"`
		Error string `json:"error,omitempty"`
	}{
		Msg:   msg,
		Error: errMsg,
	})
}

func (svc *Service) HealthCheck(c echo.Context) error {
	return ErrorResponse(c, http.StatusOK, "OK")
}

func (svc *Service) GetMemberList(c echo.Context) error {
	category := c.Param(_botServicePathKey)
	if !_validCategory[category] {
		return ErrorResponse(c, http.StatusBadGateway, fmt.Sprintf("bot '%s' not found", category))
	}

	response := model.GetMemberListResponse{}

	err := svc.repo.Tx(svc.ctx, func(txCtx context.Context) error {
		members, err := svc.repo.ListMembers(txCtx, category)
		if err != nil {
			return ErrorResponse(c, http.StatusInternalServerError, "list member error", err)
		}

		startAt, err := svc.repo.GetStartDate(txCtx, category)
		if err != nil {
			return ErrorResponse(c, http.StatusInternalServerError, "get start time error", err)
		}

		response.Members = members
		response.StartAt = startAt
		return nil
	})
	if err != nil {
		return err
	}

	return DataResponse(c, response)
}

func (svc *Service) SetMemberList(c echo.Context) error {
	category := c.Param(_botServicePathKey)
	if !_validCategory[category] {
		return ErrorResponse(c, http.StatusBadGateway, fmt.Sprintf("bot '%s' not found", category))
	}

	req := model.SetMemberListRequest{}
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "request parameters mismatch", err)
	}

	for i := range req.Members {
		req.Members[i].Service = category
	}

	sort.Slice(req.Members, func(i, j int) bool {
		return req.Members[i].Order < req.Members[j].Order
	})

	err := svc.repo.Tx(svc.ctx, func(txCtx context.Context) error {
		if err := svc.repo.ResetMembers(txCtx, category, req.Members); err != nil {
			return ErrorResponse(c, http.StatusInternalServerError, "set members error", err)
		}

		if err := svc.repo.UpdateStartDate(txCtx, category, req.StartAt); err != nil {
			return ErrorResponse(c, http.StatusInternalServerError, "set start time error", err)
		}
		return nil
	})
	if err != nil {
		return err
	}

	return svc.GetMemberList(c)
}

func (svc *Service) GetMentionMessage(c echo.Context) error {
	category := c.Param(_botServicePathKey)
	if !_validCategory[category] {
		return ErrorResponse(c, http.StatusBadGateway, fmt.Sprintf("bot '%s' not found", category))
	}

	reply, err := svc.repo.GetReplyMessage(svc.ctx, category)
	if err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "get message error", err)
	}

	return DataResponse(c, reply)
}

func (svc *Service) SetMentionMessage(c echo.Context) error {
	category := c.Param(_botServicePathKey)
	if !_validCategory[category] {
		return ErrorResponse(c, http.StatusBadGateway, fmt.Sprintf("bot '%s' not found", category))
	}

	req := model.BotMessage{}
	if err := c.Bind(&req); err != nil {
		return ErrorResponse(c, http.StatusBadRequest, "request parameters mismatch", err)
	}

	req.Service = category
	if err := svc.repo.Tx(svc.ctx, func(txCtx context.Context) error {
		return svc.repo.SetReplyMessage(txCtx, req)
	}); err != nil {
		return ErrorResponse(c, http.StatusInternalServerError, "set mention message error", err)
	}

	return svc.GetMentionMessage(c)
}
