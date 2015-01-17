package main

import (
	"bytes"
	"log"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/codegangsta/cli"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
	"github.com/learnin/goji-invited-user-signup-example/models"
)

const SMTP_CONFIG_FILE = "config/smtp.json"
const SALT = "HsE@U91Ie!8ye8ay^e87wya7Y*R%38[0(*T[9w4eut[9e"

type smtpConfig struct {
	Host     string
	Port     uint16
	Username string
	Password string
	From     string
	Subject  string
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

func main() {
	app := cli.NewApp()
	app.Name = "send invite mail"
	app.Usage = ""
	app.Action = func(c *cli.Context) {
		log.Println("招待メール送信処理を開始します。")
		defer log.Println("招待メール送信処理を終了しました。")

		action(c)
	}
	app.Run(os.Args)
}

func action(c *cli.Context) {
	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		log.Println("DB接続に失敗しました。" + err.Error())
		return
	}
	defer ds.Close()

	var inviteUsers []models.InviteUser
	if d := ds.GetDB().Where(&models.InviteUser{Status: models.STATUS_NOT_INVITED}).Find(&inviteUsers); d.Error != nil {
		log.Println("招待対象ユーザの取得に失敗しました。" + d.Error.Error())
		return
	}
	inviteUsersCount := len(inviteUsers)
	if inviteUsersCount == 0 {
		log.Println("未招待のユーザはありません。")
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
		log.Println("SMTP接続に失敗しました。" + err.Error())
		return
	}
	defer func() {
		client.Close()
		client.Quit()
	}()

	var hasError bool
	var b bytes.Buffer

	for i := 0; i < inviteUsersCount; i++ {
		inviteUser := inviteUsers[i]
		if err := ds.DoInTransaction(func(ds *helpers.DataSource) error {
			inviteUser.InviteCode = helpers.Hash(strconv.FormatInt(inviteUser.Id, 10), SALT)
			inviteUser.Status = models.STATUS_INVITED
			now := time.Now()
			inviteUser.InvitedAt = now
			inviteUser.UpdatedAt = now
			tx := ds.GetTx()
			if err := tx.Save(inviteUser).Error; err != nil {
				return err
			}
			b.Reset()
			if err := inviteMailTpl.Execute(&b, inviteUser); err != nil {
				return err
			}
			mail := helpers.Mail{
				From:    smtpCfg.From,
				To:      inviteUser.Mail,
				Subject: smtpCfg.Subject,
				Body:    b.String(),
			}
			return smtpClient.SendMail(client, mail)
		}); err != nil {
			hasError = true
			log.Println(inviteUser.Mail + " 宛のメール送信・DB更新に失敗しました。" + err.Error())
		} else {
			log.Println(inviteUser.Mail + " へ招待メールを送信しました。")
		}
	}
	if hasError {
		log.Println("メール送信・DB更新でエラーになったものがあります。")
	}
}
