package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"

	"github.com/learnin/go-multilog"
	"github.com/zenazn/goji/web"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
	"github.com/learnin/goji-invited-user-signup-example/models"
)

type ReInviteController struct {
	DS     *helpers.DataSource
	Logger *multilog.MultiLogger
}

type reInviteForm struct {
	Msg    string
	UserId string
}

func (controller *ReInviteController) ShowReInvitePage(c web.C, w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "views/index.html")
}

func (controller *ReInviteController) findInviteUserByUserId(userId string) (*models.InviteUser, error) {
	var inviteUser models.InviteUser
	if d := controller.DS.GetDB().Where(&models.InviteUser{UserId: userId}).First(&inviteUser); d.Error != nil {
		if d.RecordNotFound() {
			return nil, nil
		}
		return nil, d.Error
	}
	return &inviteUser, nil
}

func (controller *ReInviteController) validate(form reInviteForm) (bool, []string) {
	var messages []string
	if form.UserId == "" {
		messages = append(messages, "ユーザーIDを入力してください。")
	}
	if len(messages) > 0 {
		return false, messages
	}
	return true, messages
}

func (controller *ReInviteController) ReInvite(c web.C, w http.ResponseWriter, r *http.Request) {
	form := reInviteForm{}
	if err := json.NewDecoder(r.Body).Decode(&form); err != nil {
		http.Error(w, "", http.StatusBadRequest)
		return
	}
	if ok, messages := controller.validate(form); !ok {
		sendEroorResponse(w, nil, messages...)
		return
	}
	inviteUser, err := controller.findInviteUserByUserId(form.UserId)
	if err != nil {
		controller.Logger.Errorf("ユーザ検索時にエラーが発生しました。userId=%s error=%v", form.UserId, err)
		sendEroorResponse(w, err, "")
		return
	}
	if inviteUser == nil || inviteUser.IsNotInvited() {
		sendEroorResponse(w, nil, "ユーザーIDを正しく入力してください。")
		return
	}
	if inviteUser.IsSignUped() {
		sendEroorResponse(w, nil, "そのユーザーはすでに登録されています。")
		return
	}
	smtpClient := helpers.SmtpClient{
		Host:     smtpCfg.Host,
		Port:     smtpCfg.Port,
		Username: smtpCfg.Username,
		Password: smtpCfg.Password,
	}
	client, err := smtpClient.Connect()
	if err != nil {
		controller.Logger.Errorf("SMTP接続に失敗しました。 error=%v", err)
		sendEroorResponse(w, err, "")
		return
	}
	defer func() {
		client.Close()
		client.Quit()
	}()

	var b bytes.Buffer

	if err := inviteMailTpl.Execute(&b, inviteUser); err != nil {
		controller.Logger.Errorf(inviteUser.Mail+" 宛のメール作成に失敗しました。 error=%v", err)
		sendEroorResponse(w, err, "")
		return
	}
	mail := helpers.Mail{
		From:    smtpCfg.From,
		To:      inviteUser.Mail,
		Subject: smtpCfg.Subject,
		Body:    b.String(),
	}
	if err := smtpClient.SendMail(client, mail); err != nil {
		controller.Logger.Errorf(inviteUser.Mail+" 宛のメール送信に失敗しました。 error=%v", err)
		sendEroorResponse(w, err, "")
		return
	}
	controller.Logger.Info(inviteUser.Mail + " へ招待メールを再送信しました。")

	encoder := json.NewEncoder(w)
	encoder.Encode(Res{Error: false, Messages: []string{}})
}
