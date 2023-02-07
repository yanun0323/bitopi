package app

import (
	"bitopi/internal/model"
	"bitopi/internal/service"
	"bitopi/internal/util"
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/robfig/cron/v3"
	"github.com/spf13/viper"
)

func Run() {
	e := echo.New()
	e.Logger.SetLevel(4)

	rateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20))
	m := []echo.MiddlewareFunc{rateLimiter}
	router := e.Group("", m...)
	router.GET("/debug", service.DebugHandler, m...)
	router.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Msg string `json:"message"`
		}{
			Msg: "OK",
		})
	})

	if err := setupRouters(router); err != nil {
		panic(fmt.Sprintf("setup routers error, %+v", err))
	}

	e.Start(":8001")
}

func setupRouters(router *echo.Group) error {
	svc, err := service.New()
	if err != nil {
		return err
	}

	if err := setBot(router, svc, service.SlackBotOption{
		Name:             "pm",
		Token:            viper.GetString("pm.token"),
		DefaultStartDate: util.NewDate(2022, 11, 27),
		DefaultMemberList: []model.Member{
			{UserID: "U02223HG26L", UserName: "Rafeni"},
			{UserID: "U01THK4U2MD", UserName: "Momo"},
		},
		DefaultReplyMessage: "請稍候片刻，%s Support PM %s 將盡快為您服務 :smiling_face_with_3_hearts:",
	}); err != nil {
		return err
	}

	if err := setBot(router, svc, service.SlackBotOption{
		Name:             "rails",
		Token:            viper.GetString("rails.token"),
		DefaultStartDate: util.NewDate(2022, 11, 6),
		DefaultMemberList: []model.Member{
			{UserID: "U0156SRG9GW", UserName: "Gmi"},
			{UserID: "UKL1DAL4E", UserName: "Barry"},
			{UserID: "U0328D2JE8H", UserName: "Kurt"},
			{UserID: "UQTAPAF2T", UserName: "Kevin"},
			{UserID: "U041HD3AQ3D", UserName: "Eric"},
			{UserID: "U01GTQ8K52P", UserName: "Yuan"},
		},
		DefaultReplyMessage: "請稍候片刻，本週茅房廁紙 %s 會盡快為您服務 :smiling_face_with_3_hearts:",
	}); err != nil {
		return err
	}

	if err := setBot(router, svc, service.SlackBotOption{
		Name:             "devops",
		Token:            viper.GetString("devops.token"),
		DefaultStartDate: util.NewDate(2022, 10, 23),
		DefaultMemberList: []model.Member{
			{UserID: "U03FDTNPWBW", UserName: "Lawrence"},
			{UserID: "U03RQKWLG8Z", UserName: "Tina"},
			{UserID: "U01A7LEG1CZ", UserName: "Harlan"},
		},
		DefaultReplyMessage: "請稍候片刻，本週猛哥/猛姐會盡快為您服務 :smiling_face_with_3_hearts:\nBito EX/Pro: %s\nMeta: %s",
		DefaultMultiMember:  true,
	}); err != nil {
		return err
	}

	if err := setBot(router, svc, service.SlackBotOption{
		Name:             "maid",
		Token:            viper.GetString("maid.token"),
		DefaultStartDate: util.NewDate(2022, 9, 25),
		DefaultMemberList: []model.Member{
			{UserID: "U032TJB1PE1", UserName: "Yanun"},
			{UserID: "U03ECC8Q61E", UserName: "Howard"},
			{UserID: "U031SSN3QDT", UserName: "Kai"},
			{UserID: "U01QCKG7529", UserName: "Vic"},
			{UserID: "U036V8WPXDY", UserName: "Victor"},
			{UserID: "U03MWAJDBV3", UserName: "Luki"},
		},
		DefaultReplyMessage: "請稍候片刻，本週女僕 %s 會盡快為您服務 :smiling_face_with_3_hearts:",
	}); err != nil {
		return err
	}

	if err := setBot(router, svc, service.SlackBotOption{
		Name:             "test",
		Token:            viper.GetString("test.token"),
		DefaultStartDate: util.NewDate(2023, 1, 22),
		DefaultMemberList: []model.Member{
			{UserID: "U032TJB1PE1", UserName: "Yanun"},
			{UserID: "U032TJB1PE1", UserName: "Yanun"},
			{UserID: "U032TJB1PE1", UserName: "Yanun"},
		},
		DefaultReplyMessage: "測試訊息，今日值日生 %s :smiling_face_with_3_hearts:",
	}); err != nil {
		return err
	}

	return nil
}

func setBot(router *echo.Group, svc service.Service, opt service.SlackBotOption) error {
	bot := service.NewBot(svc, opt)
	action := service.NewInteraction(bot)

	router.POST(fmt.Sprintf("/%s", bot.Name), bot.Handler)
	router.POST(fmt.Sprintf("/%s/action", bot.Name), action.Handler)

	if err := setupCron(bot, service.WeeklyNotifierOpt{}); err != nil {
		return err
	}

	return nil
}

func setupCron(bot service.SlackBot, opt service.WeeklyNotifierOpt) error {
	job := service.NewWeeklyJob(bot, opt)
	c := cron.New(cron.WithSeconds())
	_, err := c.AddJob("TZ=Asia/Taipei 0 0 9 ? * 0", job)
	if err != nil {
		return err
	}
	return nil
}
