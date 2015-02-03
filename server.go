package main

import (
	"log"
	"net/http"

	"github.com/zenazn/goji"
	"github.com/zenazn/goji/graceful"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/learnin/goji-invited-user-signup-example/controllers"
	"github.com/learnin/goji-invited-user-signup-example/helpers"
)

func main() {
	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		panic(err)
	}

	signUp := web.New()
	goji.Handle("/signup/*", signUp)
	signUp.Use(middleware.SubRouter)
	signUpController := controllers.SignUpController{DS: &ds}
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
			log.Println(err)
		}
	})

	goji.Serve()
}
