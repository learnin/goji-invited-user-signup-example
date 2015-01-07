package main

import (
	"fmt"
	"log"
	"net/smtp"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/cli"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
	"github.com/learnin/goji-invited-user-signup-example/models"
)

const SALT = "HsE@U91Ie!8ye8ay^e87wya7Y*R%38[0(*T[9w4eut[9e"
const SMTP_CONFIG_FILE = "config/smtp.json"

type smtpConfig struct {
	Host     string
	Port     uint16
	Username string
	Password string
	From     string
}

var smtpCfg smtpConfig

func init() {
	jsonHelper := helpers.Json{}
	if err := jsonHelper.UnmarshalJsonFile(SMTP_CONFIG_FILE, &smtpCfg); err != nil {
		log.Fatalln(err)
	}
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
			return sendMail(inviteUser)
		}); err != nil {
			// FIXME エラーが発生してもスキップするようにする
			return err
		}
	}
	return nil
}

// FIXME 接続は1回にして、Reset, Mail, Rcpt, Data, w.Closeを宛先文回す。
func sendMail(inviteUser models.InviteUser) error {
	c, err := smtp.Dial(fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port))
	if err != nil {
		return err
	}
	defer c.Close()

	// auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)

	if err := c.Hello("localhost"); err != nil {
		return err
	}
	// if ok, _ := c.Extension("AUTH"); ok {
	// 	if err := c.Auth(auth); err != nil {
	// 		return err
	// 	}
	// }

	c.Mail(smtpCfg.From)
	if err := c.Rcpt(inviteUser.Mail); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	// FIXME エンコード
	msg := "Content-Type: text/plain; charset=ISO-2022-JP\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"From: " + smtpCfg.From + "\r\n" +
		"To: " + inviteUser.Mail + "\r\n" +
		"Subject: test\r\n" +
		"\r\n" +
		"テストメールです"
	if _, err = w.Write([]byte(msg)); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}
	return c.Quit()
}
