package controllers

import (
	"net/http"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji/web"

	"github.com/learnin/goji-invited-user-signup-example/models"
)

type SignUpController struct {
}

type UserForm struct {
	HashKey         string
	Msg             string
	UserId          string
	Password        string
	ConfirmPassword string
	LastName        string
	FirstName       string
	Mail            string
}

var signupTpl = pongo2.Must(pongo2.FromFile("views/signup.tpl"))
var completeTpl = pongo2.Must(pongo2.FromFile("views/complete.tpl"))

func request2UserForm(r *http.Request) UserForm {
	var form UserForm
	form.UserId = r.FormValue("userId")
	form.Password = r.FormValue("password")
	form.ConfirmPassword = r.FormValue("confirmPassword")
	form.LastName = r.FormValue("lastName")
	form.FirstName = r.FormValue("firstName")
	form.Mail = r.FormValue("mail")
	form.HashKey = r.FormValue("hashKey")
	return form
}

func (controller *SignUpController) ShowSignupPage(c web.C, w http.ResponseWriter, r *http.Request) {
	hashKey := c.URLParams["hashKey"]
	if hashKey == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	var form UserForm
	form.HashKey = hashKey
	controller.renderSignupPage(c, w, r, form)
}

func (controller *SignUpController) ShowCompletePage(c web.C, w http.ResponseWriter, r *http.Request) {
	err := completeTpl.ExecuteWriter(pongo2.Context{"": ""}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (controller *SignUpController) renderSignupPage(c web.C, w http.ResponseWriter, r *http.Request, form UserForm) {
	err := signupTpl.ExecuteWriter(pongo2.Context{"form": form}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (controller *SignUpController) userForm2User(form UserForm) models.User {
	var user models.User
	user.UserId = form.UserId
	user.Password = form.Password
	user.LastName = form.LastName
	user.FirstName = form.FirstName
	user.Mail = form.Mail
	user.HashKey = form.HashKey
	return user
}

func (controller *SignUpController) SignUp(c web.C, w http.ResponseWriter, r *http.Request) {
	form := request2UserForm(r)
	if form.HashKey == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	user := controller.userForm2User(form)
	if ok, msg := controller.validate(form, user); !ok {
		msg = strings.Replace(msg, "\n", "<br/>", -1)
		form.Msg = msg
		controller.renderSignupPage(c, w, r, form)
		return
	}
	if !user.ValidateHashKey() {
		form.Msg = "ユーザーIDを正しく入力してください。"
		controller.renderSignupPage(c, w, r, form)
		return
	}
	err := user.AddUser()
	if err != nil {
		switch err.(type) {
		case models.AlreadyExistError:
			form.Msg = "そのユーザーはすでに登録されています。"
			controller.renderSignupPage(c, w, r, form)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.Redirect(w, r, "/signup/complete", http.StatusFound)
}

func (controller *SignUpController) validate(form UserForm, user models.User) (bool, string) {
	_, msg := user.Validate()
	if form.ConfirmPassword == "" {
		msg += "パスワード(確認)を入力してください。\n"
	} else if form.Password != "" && form.Password != form.ConfirmPassword {
		msg += "パスワードとパスワード(確認)が一致していません。\n"
	}
	if msg != "" {
		return false, msg
	}
	return true, ""
}
