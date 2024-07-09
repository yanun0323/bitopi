package util

import "fmt"

const (
	_statusOK   = "OK"
	_statusFail = "FAIL"
)

type Response struct {
	Status string      `json:"status"`
	Msg    string      `json:"msg"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

func NewMsgResponse(msg string) Response {
	return Response{
		Status: _statusOK,
		Msg:    msg,
	}
}

func NewErrorResponse(msg string, errs ...error) Response {
	if len(errs) == 0 || errs[0] == nil {
		return Response{
			Status: _statusFail,
			Msg:    msg,
		}
	}
	return Response{
		Status: _statusFail,
		Msg:    msg,
		Error:  fmt.Sprintf("%s", errs[0]),
	}
}

func NewDataResponse(msg string, data interface{}) Response {
	return Response{
		Status: _statusOK,
		Msg:    msg,
		Data:   data,
	}
}
