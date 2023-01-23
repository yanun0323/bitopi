package app

import (
	"bitopi/internal/service"
	"bitopi/internal/util"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/spf13/viper"
	"github.com/yanun0323/pkg/logs"
)

func Run() {
	l := logs.New("bito_pi", uint8(viper.GetInt("log.level")))
	e := echo.New()
	e.Logger.SetLevel(4)

	svc, err := service.New()
	if err != nil {
		l.Fatalf("create service failed %s", err)
		return
	}

	rateLimiter := middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(20))
	m := []echo.MiddlewareFunc{rateLimiter}
	router := e.Group("", m...)
	router.GET("/healthz", func(c echo.Context) error {
		return c.JSON(http.StatusOK, struct {
			Msg string `json:"message"`
		}{
			Msg: "OK",
		})
	}, m...)
	router.GET("/debug", svc.DebugService, m...)

	pmBot := service.NewBot("pm", svc, service.SlackBotOption{
		DefaultStartDate: util.NewDate(2022, 11, 27),
		DefaultMemberList: []string{
			"<@U02223HG26L>", /* Rafeni */
			"<@U01THK4U2MD>", /* Momo */
		},
		Token:          viper.GetString("pm.token"),
		ReplyMsgFormat: "請稍候片刻，%s Support PM %s 將盡快為您服務 :smiling_face_with_3_hearts:",
	})

	railsBot := service.NewBot("rails", svc, service.SlackBotOption{
		DefaultStartDate: util.NewDate(2022, 11, 6),
		DefaultMemberList: []string{
			"<@U0156SRG9GW>", /* Gmi */
			"<@UKL1DAL4E>",   /* Barry */
			"<@U0328D2JE8H>", /* Kurt */
			"<@UQTAPAF2T>",   /* Kevin */
			"<@U041HD3AQ3D>", /* Eric */
			"<@U01GTQ8K52P>", /* Yuan */
		},
		Token:          viper.GetString("rails.token"),
		ReplyMsgFormat: "請稍候片刻，本週茅房廁紙 %s 會盡快為您服務 :smiling_face_with_3_hearts:",
	})

	devopsBot := service.NewBot("devops", svc, service.SlackBotOption{
		DefaultStartDate: util.NewDate(2022, 10, 23),
		DefaultMemberList: []string{
			"<@U03FDTNPWBW>", /* Lawrence */
			"<@U03RQKWLG8Z>", /* Tina */
			"<@U01A7LEG1CZ>", /* Harlan */
		},
		Token:          viper.GetString("devops.token"),
		ReplyMsgFormat: "請稍候片刻，本週猛哥/猛姐會盡快為您服務 :smiling_face_with_3_hearts:\nBito EX/Pro: %s\nMeta: %s",
		IsMultiMember:  true,
	})

	maidBot := service.NewBot("maid", svc, service.SlackBotOption{
		DefaultStartDate: util.NewDate(2022, 9, 25),
		DefaultMemberList: []string{
			"<@U032TJB1PE1>", /* Yanun */
			"<@U03ECC8Q61E>", /* Howard */
			"<@U031SSN3QDT>", /* Kai */
			"<@U01QCKG7529>", /* Vic */
			"<@U036V8WPXDY>", /* Victor */
			"<@U03MWAJDBV3>", /* Luki */
		},
		Token:          viper.GetString("maid.token"),
		ReplyMsgFormat: "請稍候片刻，本週女僕 %s 會盡快為您服務 :smiling_face_with_3_hearts:",
	})

	testBot := service.NewBot("test", svc, service.SlackBotOption{
		DefaultStartDate: util.NewDate(2023, 1, 22),
		DefaultMemberList: []string{
			"<@U032TJB1PE1>", /* Yanun */
			"<@U03ECC8Q61E>", /* Howard */
			"<@U031SSN3QDT>", /* Kai */
			"<@U01QCKG7529>", /* Vic */
			"<@U036V8WPXDY>", /* Victor */
			"<@U03MWAJDBV3>", /* Luki */
		},
		Token:          viper.GetString("test.token"),
		ReplyMsgFormat: "測試訊息 %s :smiling_face_with_3_hearts:",
	})

	router.POST("/pm", pmBot.Handler)
	router.POST("/rails-hi", railsBot.Handler)
	router.POST("/devops-bro", devopsBot.Handler)
	router.POST("/backend-maid", maidBot.Handler)
	router.POST("/test", testBot.Handler)
	// router.POST("/backend-maid/command", svc.MaidCommandHandler, m...)

	e.Start(":8001")
}
