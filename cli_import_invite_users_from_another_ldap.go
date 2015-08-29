package main

import (
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/jinzhu/gorm"
	"github.com/learnin/go-multilog"
	"github.com/mattn/go-colorable"
	"github.com/mavricknz/ldap"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
	"github.com/learnin/goji-invited-user-signup-example/models"
)

const LDAP_CONFIG_FILE = "config/another_ldap.json"
const LOG_DIR = "log"
const LOG_FILE = LOG_DIR + "/cli_import_invite_users_from_another_ldap.log"

type ldapConfig struct {
	Host         string
	Port         uint16
	BindDn       string
	BindPassword string
	BaseDn       string
	Filter       string
}

type ldapUser struct {
	Uid       string
	GivenName string
	Sn        string
	Mail      string
}

var ldapCfg ldapConfig
var log *multilog.MultiLogger

func init() {
	jsonHelper := helpers.Json{}
	if err := jsonHelper.UnmarshalJsonFile(LDAP_CONFIG_FILE, &ldapCfg); err != nil {
		panic(err)
	}
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
	app.Name = "import-invite-users-from-another-ldap"
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
	app.Usage = "import invite users from another LDAP."
	app.Action = func(c *cli.Context) {
		log.Info("LDAPサーバからの招待ユーザインポート処理を開始します。")
		defer log.Info("LDAPサーバからの招待ユーザインポート処理を終了しました。")

		action(c)
	}
	app.Run(os.Args)
}

func getUsersFromAnotherLdap() ([]ldapUser, error) {
	l := ldap.NewLDAPConnection(ldapCfg.Host, ldapCfg.Port)
	if err := l.Connect(); err != nil {
		return nil, err
	}
	defer l.Close()
	if ldapCfg.BindDn != "" {
		if err := l.Bind(ldapCfg.BindDn, ldapCfg.BindPassword); err != nil {
			return nil, err
		}
	}
	searchRequest := ldap.NewSearchRequest(
		ldapCfg.BaseDn,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		ldapCfg.Filter,
		[]string{"uid", "givenName", "sn", "mail"},
		nil)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return nil, err
	}
	ldapUsers := make([]ldapUser, len(sr.Entries))
	for _, entry := range sr.Entries {
		user := ldapUser{
			Uid:       entry.GetAttributeValue("uid"),
			GivenName: entry.GetAttributeValue("givenName"),
			Sn:        entry.GetAttributeValue("sn"),
			Mail:      entry.GetAttributeValue("mail"),
		}
		ldapUsers = append(ldapUsers, user)
	}
	return ldapUsers, nil
}

func existsInviteUserByUserId(ds helpers.DataSource, userId string) (bool, error) {
	var count int
	if d := ds.GetDB().Model(models.InviteUser{}).Where(&models.InviteUser{UserId: userId}).Count(&count); d.Error != nil {
		return false, d.Error
	}
	return count > 0, nil
}

func action(c *cli.Context) {
	isVerbose := c.Bool("verbose")

	ldapUsers, err := getUsersFromAnotherLdap()
	if err != nil {
		log.Error("LDAPサーバからのユーザ取得に失敗しました。" + err.Error())
		return
	}
	if len(ldapUsers) == 0 {
		log.Error("LDAPサーバにユーザが存在しません。" + LDAP_CONFIG_FILE + " を確認ください。")
		return
	}

	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		log.Error("DB接続に失敗しました。" + err.Error())
		return
	}
	defer ds.Close()

	if isVerbose {
		ds.LogMode(true)
	}

	var hasError bool
	existsUserCount := 0

	for _, user := range ldapUsers {
		if user.Uid == "" {
			continue
		}
		if existsInviteUser, err := existsInviteUserByUserId(ds, user.Uid); err != nil {
			log.Error("招待ユーザの取得に失敗しました。" + err.Error())
			return
		} else if existsInviteUser {
			existsUserCount++
			continue
		}
		now := time.Now()
		inviteUser := models.InviteUser{
			UserId:    user.Uid,
			LastName:  user.Sn,
			FirstName: user.GivenName,
			Mail:      user.Mail,
			Status:    models.STATUS_NOT_INVITED,
			CreatedAt: now,
			UpdatedAt: now,
		}
		if err := ds.DoInTransaction(func(tx *gorm.DB) error {
			return tx.Create(&inviteUser).Error
		}); err != nil {
			hasError = true
			log.Error("UserId: " + inviteUser.UserId + " ユーザのDB登録に失敗しました。" + err.Error())
		} else {
			log.Info("UserId: " + inviteUser.UserId + " ユーザをDBへ登録しました。")
		}
	}
	if hasError {
		log.Error("DB登録でエラーになったものがあります。")
	}
	if existsUserCount == len(ldapUsers) {
		log.Info("未登録のユーザはありませんでした。")
	}
}
