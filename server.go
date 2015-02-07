package main

import (
	"net/http"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/learnin/go-multilog"
	"github.com/mattn/go-colorable"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/learnin/goji-invited-user-signup-example/controllers"
	"github.com/learnin/goji-invited-user-signup-example/helpers"
)

const LOG_DIR = "log"
const LOG_FILE = LOG_DIR + "/server.log"

func main() {
	var log *multilog.MultiLogger
	if fi, err := os.Stat(LOG_DIR); os.IsNotExist(err) {
		if err := os.MkdirAll(LOG_DIR, 0755); err != nil {
			panic(err)
		}
	} else {
		if !fi.IsDir() {
			panic("ログディレクトリ " + LOG_DIR + " はディレクトリではありません。")
		}
	}
	logf, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		panic(err)
	}
	stdOutLogrus := logrus.New()
	stdOutLogrus.Out = colorable.NewColorableStdout()
	fileLogrus := logrus.New()
	fileLogrus.Out = logf
	fileLogrus.Formatter = &logrus.TextFormatter{DisableColors: true}
	log = multilog.New(stdOutLogrus, fileLogrus)

	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		panic(err)
	}

	signUp := web.New()
	goji.Handle("/signup/*", signUp)
	signUp.Use(middleware.SubRouter)
	signUpController := controllers.SignUpController{DS: &ds, Logger: log}
	signUp.Post("/execute", signUpController.SignUp)
	signUp.Get("/:inviteCode", signUpController.ShowSignupPage)

	assets := web.New()
	assets.Get("/assets/*", http.StripPrefix("/assets/", http.FileServer(http.Dir("assets"))))
	http.Handle("/assets/", assets)

	views := web.New()
	views.Get("/views/*", http.StripPrefix("/views/", http.FileServer(http.Dir("views"))))
	http.Handle("/views/", views)

	graceful.PostHook(func() {
		if err := ds.Close(); err != nil {
			log.Errorln(err)
		}
		logf.Close()
	})

	goji.Serve()
}
