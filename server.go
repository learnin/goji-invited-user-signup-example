package main

import (
	"log"

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
		log.Fatalln(err)
	}

	signUp := web.New()
	goji.Handle("/signup/*", signUp)
	signUp.Use(middleware.SubRouter)
	signUpController := controllers.SignUpController{DS: &ds}
	signUp.Get("/:hashKey", signUpController.ShowSignupPage)
	signUp.Post("/execute", signUpController.SignUp)
	signUp.Get("/complete", signUpController.ShowCompletePage)

	graceful.PostHook(func() {
		if err := ds.Close(); err != nil {
			log.Fatalln(err)
		}
	})

	goji.Serve()
}
