package main

import (
	"bytes"
	"os"
	"strconv"
	"text/template"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jinzhu/gorm"
	"github.com/learnin/go-multilog"
	"github.com/mattn/go-colorable"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
	"github.com/learnin/goji-invited-user-signup-example/models"
)

const SMTP_CONFIG_FILE = "config/smtp.json"
const LOG_DIR = "log"
const LOG_FILE = LOG_DIR + "/cli_send_invite_mail.log"
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
var log *multilog.MultiLogger

func init() {
	jsonHelper := helpers.Json{}
	if err := jsonHelper.UnmarshalJsonFile(SMTP_CONFIG_FILE, &smtpCfg); err != nil {
		panic(err)
	}
	inviteMailTpl = template.Must(template.ParseFiles("config/invite_mail.tpl"))
}

func main() {
	if fi, err := os.Stat(LOG_DIR); os.IsNotExist(err) {
		if err := os.MkdirAll(LOG_DIR, 0755); err != nil {
			panic(err)
		}
	} else {
		if !fi.IsDir() {
			panic("ログディレクトリ " + LOG_DIR + " はディレクトリではありません。")
		}
	}
	if logf, err := os.OpenFile(LOG_FILE, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err != nil {
		panic(err)
	} else {
		defer logf.Close()
		stdOutLogrus := logrus.New()
		stdOutLogrus.Out = colorable.NewColorableStdout()
		fileLogrus := logrus.New()
		fileLogrus.Out = logf
		fileLogrus.Formatter = &logrus.TextFormatter{DisableColors: true}
		log = multilog.New(stdOutLogrus, fileLogrus)
	}

	app := cli.NewApp()
	app.Name = "send-invite-mail"
	app.Version = "0.0.1"
	app.Author = "Manabu Inoue"
	app.Email = ""
	app.HideVersion = true
	app.EnableBashCompletion = true
	app.Flags = []cli.Flag{
		cli.BoolFlag{
			Name:  "verbose, v",
			Usage: "verbose mode. a lot more information output",
		},
		cli.BoolFlag{
			Name:  "version, V",
			Usage: "print the version",
		},
	}
	app.Usage = "send invite mail for signup."
	app.Action = func(c *cli.Context) {
		log.Info("招待メール送信処理を開始します。")
		defer log.Info("招待メール送信処理を終了しました。")

		action(c)
	}
	app.Run(os.Args)
}

func action(c *cli.Context) {
	isVerbose := c.Bool("verbose")

	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		log.Error("DB接続に失敗しました。" + err.Error())
		return
	}
	defer ds.Close()

	if isVerbose {
		ds.LogMode(true)
	}

	var inviteUsers []models.InviteUser
	if d := ds.GetDB().Where(&models.InviteUser{Status: models.STATUS_NOT_INVITED}).Find(&inviteUsers); d.Error != nil {
		log.Error("招待対象ユーザの取得に失敗しました。" + d.Error.Error())
		return
	}
	inviteUsersCount := len(inviteUsers)
	if inviteUsersCount == 0 {
		log.Info("未招待のユーザはありません。")
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
		log.Error("SMTP接続に失敗しました。" + err.Error())
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
		if err := ds.DoInTransaction(func(tx *gorm.DB) error {
			inviteUser.InviteCode = helpers.Hash(strconv.FormatInt(inviteUser.Id, 10), SALT)
			inviteUser.Status = models.STATUS_INVITED
			now := time.Now()
			inviteUser.InvitedAt = now
			inviteUser.UpdatedAt = now
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
			log.Error(inviteUser.Mail + " 宛のメール送信・DB更新に失敗しました。" + err.Error())
		} else {
			log.Info(inviteUser.Mail + " へ招待メールを送信しました。")
		}
	}
	if hasError {
		log.Error("メール送信・DB更新でエラーになったものがあります。")
	}
}
