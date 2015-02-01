package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/zenazn/goji/web"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
	"github.com/learnin/goji-invited-user-signup-example/models"
)

const DEBUG = true

type SignUpController struct {
	DS *helpers.DataSource
}

type UserForm struct {
	InviteCode      string
	Msg             string
	UserId          string
	Password        string
	ConfirmPassword string
}

type Res struct {
	Error        bool     `json:"error"`
	Messages     []string `json:"messages"`
	DebugMessage string   `json:"debugMessage"`
}

func snedEroorResponse(w http.ResponseWriter, e error, messages ...string) {
	if messages[0] == "" {
		messages = []string{"システムエラーが発生しました。"}
	}
	res := Res{
		Error:    true,
		Messages: messages,
	}
	if DEBUG && e != nil {
		res.DebugMessage = e.Error()
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(res)
}

func (controller *SignUpController) ShowSignupPage(c web.C, w http.ResponseWriter, r *http.Request) {
	inviteCode := c.URLParams["inviteCode"]
	inviteUser, err := controller.findInviteUserByInviteCode(inviteCode)
	if err != nil {
		http.Error(w, "システムエラーが発生しました。", 500)
		return
	}
	if inviteUser == nil || inviteUser.IsNotInvited() {
		http.Error(w, "URLが誤っています。", 404)
		return
	}
	if inviteUser.IsSignUped() {
		http.Error(w, "すでに登録されています。", 200)
		return
	}
	http.ServeFile(w, r, "views/index.html")
}

func (controller *SignUpController) userForm2User(form UserForm) models.User {
	var user models.User
	user.UserId = form.UserId
	user.Password = form.Password
	user.InviteCode = form.InviteCode
	return user
}

func (controller *SignUpController) findInviteUserByInviteCode(inviteCode string) (*models.InviteUser, error) {
	var inviteUser models.InviteUser
	if d := controller.DS.GetDB().Where(&models.InviteUser{InviteCode: inviteCode}).First(&inviteUser); d.Error != nil {
		if d.RecordNotFound() {
			return nil, nil
		}
		return nil, d.Error
	}
	return &inviteUser, nil
}

func (controller *SignUpController) findInviteUserByUserId(userId string) (*models.InviteUser, error) {
	var inviteUser models.InviteUser
	if d := controller.DS.GetDB().Where(&models.InviteUser{UserId: userId}).First(&inviteUser); d.Error != nil {
		if d.RecordNotFound() {
			return nil, nil
		}
		return nil, d.Error
	}
	return &inviteUser, nil
}

func (controller *SignUpController) validate(form UserForm, user models.User) (bool, []string) {
	_, messages := user.Validate()
	if form.ConfirmPassword == "" {
		messages = append(messages, "パスワード(確認)を入力してください。")
	} else if form.Password != "" && form.Password != form.ConfirmPassword {
		messages = append(messages, "パスワードとパスワード(確認)が一致していません。")
	}
	if len(messages) > 0 {
		return false, messages
	}
	return true, messages
}

func (controller *SignUpController) SignUp(c web.C, w http.ResponseWriter, r *http.Request) {
	form := UserForm{}
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if form.InviteCode == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	user := controller.userForm2User(form)
	if ok, messages := controller.validate(form, user); !ok {
		snedEroorResponse(w, nil, messages...)
		return
	}
	inviteUser, err := controller.findInviteUserByUserId(user.UserId)
	if err != nil {
		snedEroorResponse(w, err, "")
		return
	}
	if inviteUser == nil || inviteUser.InviteCode != user.InviteCode || inviteUser.IsNotInvited() {
		snedEroorResponse(w, nil, "ユーザーIDを正しく入力してください。")
		return
	}
	if inviteUser.IsSignUped() {
		snedEroorResponse(w, nil, "そのユーザーはすでに登録されています。")
		return
	}
	user.LastName = inviteUser.LastName
	user.FirstName = inviteUser.FirstName
	user.Mail = inviteUser.Mail
	err = user.AddUser(controller.DS, inviteUser)
	if err != nil {
		switch err.(type) {
		case models.AlreadyExistError:
			snedEroorResponse(w, nil, "そのユーザーはすでに登録されています。")
		default:
			snedEroorResponse(w, err, "")
		}
		return
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(Res{Error: false, Messages: []string{}})
}
