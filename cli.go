package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/codegangsta/cli"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
	"github.com/learnin/goji-invited-user-signup-example/models"
)

const SALT = "HsE@U91Ie!8ye8ay^e87wya7Y*R%38[0(*T[9w4eut[9e"

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
	smtpUtil := helpers.SmtpUtil{}
	client, err := smtpUtil.Connect()
	if err != nil {
		return err
	}
	defer client.Close()

	var e error

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
			return smtpUtil.SendMail(client, inviteUser.Mail)
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
