package app

import (
	"bitopi/internal/model"
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
		Token:            viper.GetString("pm.token"),
		DefaultStartDate: util.NewDate(2022, 11, 27),
		DefaultMemberList: []model.Member{
			{UserID: "U02223HG26L", UserName: "Rafeni"},
			{UserID: "U01THK4U2MD", UserName: "Momo "},
		},
		DefaultReplyMessage: "請稍候片刻，%s Support PM %s 將盡快為您服務 :smiling_face_with_3_hearts:",
	})

	railsBot := service.NewBot("rails", svc, service.SlackBotOption{
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
	})

	devopsBot := service.NewBot("devops", svc, service.SlackBotOption{
		Token:            viper.GetString("devops.token"),
		DefaultStartDate: util.NewDate(2022, 10, 23),
		DefaultMemberList: []model.Member{
			{UserID: "U03FDTNPWBW", UserName: "Lawrence"},
			{UserID: "U03RQKWLG8Z", UserName: "Tina"},
			{UserID: "U01A7LEG1CZ", UserName: "Harlan"},
		},
		DefaultReplyMessage: "請稍候片刻，本週猛哥/猛姐會盡快為您服務 :smiling_face_with_3_hearts:\nBito EX/Pro: %s\nMeta: %s",
		DefaultMultiMember:  true,
	})

	maidBot := service.NewBot("maid", svc, service.SlackBotOption{
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
	})

	maidAction := service.NewAction("maid", maidBot)

	testBot := service.NewBot("test", svc, service.SlackBotOption{
		Token:            viper.GetString("test.token"),
		DefaultStartDate: util.NewDate(2023, 1, 22),
		DefaultMemberList: []model.Member{
			{UserID: "U032TJB1PE1", UserName: "Yanun"},
			{UserID: "U03ECC8Q61E", UserName: "Howard"},
			{UserID: "U031SSN3QDT", UserName: "Kai"},
			{UserID: "U01QCKG7529", UserName: "Vic"},
			{UserID: "U036V8WPXDY", UserName: "Victor"},
			{UserID: "U03MWAJDBV3", UserName: "Luki"},
		},
		DefaultReplyMessage: "測試訊息 %s :smiling_face_with_3_hearts:",
	})

	router.POST("/pm", pmBot.Handler)
	router.POST("/rails-hi", railsBot.Handler)
	router.POST("/devops-bro", devopsBot.Handler)
	router.POST("/backend-maid", maidBot.Handler)
	router.POST("/test", testBot.Handler)
	// router.POST("/backend-maid/command", svc.MaidCommandHandler, m...)

	router.POST("/backend-maid/action", maidAction.Handler)

	e.Start(":8001")
}
