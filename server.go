package main

import (
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"github.com/zenazn/goji/web/middleware"

	"github.com/learnin/goji-invited-user-signup-example/controllers"
)

func main() {
	signUp := web.New()
	goji.Handle("/signup/*", signUp)
	signUp.Use(middleware.SubRouter)
	var signUpController controllers.SignUpController
	signUp.Get("/:hashKey", signUpController.ShowSignupPage)
	signUp.Post("/execute", signUpController.SignUp)
	signUp.Get("/complete", signUpController.ShowCompletePage)

	goji.Serve()
}
