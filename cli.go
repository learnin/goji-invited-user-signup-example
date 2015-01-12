package main

import (
	"bytes"
	"fmt"
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
		log.Fatalln(err)
	}
	inviteMailTpl = template.Must(template.ParseFiles("config/invite_mail.tpl"))
}

func main() {
	app := cli.NewApp()
	app.Name = "greet"
	app.Usage = "fight the loneliness!"
	app.Action = func(c *cli.Context) {
		if err := action(c); err != nil {
			log.Fatalln(err)
		}
	}

	app.Run(os.Args)
}

func action(c *cli.Context) error {
	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		return err
	}
	defer ds.Close()

	var inviteUsers []models.InviteUser
	if d := ds.GetDB().Where(&models.InviteUser{Status: models.STATUS_NOT_INVITED}).Find(&inviteUsers); d.Error != nil {
		return d.Error
	}
	inviteUsersCount := len(inviteUsers)
	if inviteUsersCount == 0 {
		fmt.Println("未招待のユーザはありません。")
		return nil
	}
	smtpClient := helpers.SmtpClient{
		Host:     smtpCfg.Host,
		Port:     smtpCfg.Port,
		Username: smtpCfg.Username,
		Password: smtpCfg.Password,
	}
	client, err := smtpClient.Connect()
	if err != nil {
		return err
	}
	defer client.Close()

	var e error
	var b bytes.Buffer

	for i := 0; i < inviteUsersCount; i++ {
		if err := ds.DoInTransaction(func(ds *helpers.DataSource) error {
			inviteUser := inviteUsers[i]
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
			// FIXME エラーが発生してもスキップするようにする
			e = err
			break
		}
	}
	if e != nil {
		client.Quit()
		return e
	}
	return client.Quit()
}
