package controllers

import (
	"encoding/json"
	"net/http"
	"text/template"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
)

const DEBUG = true
const SMTP_CONFIG_FILE = "config/smtp.json"

type smtpConfig struct {
	Host     string
	Port     uint16
	Username string
	Password string
	From     string
	Subject  string
}

type response struct {
	Error        bool     `json:"error"`
	Messages     []string `json:"messages"`
	DebugMessage string   `json:"debugMessage"`
}

var smtpCfg smtpConfig
var inviteMailTpl *template.Template

func init() {
	jsonHelper := helpers.Json{}
	if err := jsonHelper.UnmarshalJsonFile(SMTP_CONFIG_FILE, &smtpCfg); err != nil {
		panic(err)
	}
	inviteMailTpl = template.Must(template.ParseFiles("config/invite_mail.tpl"))
}

func sendEroorResponse(w http.ResponseWriter, e error, messages ...string) {
	if messages[0] == "" {
		messages = []string{"システムエラーが発生しました。"}
	}
	res := response{
		Error:    true,
		Messages: messages,
	}
	if DEBUG && e != nil {
		res.DebugMessage = e.Error()
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(res)
}
