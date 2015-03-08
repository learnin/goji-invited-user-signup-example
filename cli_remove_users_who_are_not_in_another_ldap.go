package main

import (
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/codegangsta/cli"
	"github.com/learnin/go-multilog"
	"github.com/mattn/go-colorable"
	"github.com/mavricknz/ldap"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
	"github.com/learnin/goji-invited-user-signup-example/models"
)

const LDAP_CONFIG_FILE = "config/ldap.json"
const ANOTHER_LDAP_CONFIG_FILE = "config/another_ldap.json"
const LOG_DIR = "log"
const LOG_FILE = LOG_DIR + "/cli_remove_users_who_are_not_in_another_ldap.log"

type ldapConfig struct {
	Host         string
	Port         uint16
	BindDn       string
	BindPassword string
	BaseDn       string
	Filter       string
	Verbose      bool
}

type ldapUser struct {
	Uid       string
	GivenName string
	Sn        string
	Mail      string
}

var ldapCfg ldapConfig
var anotherLdapCfg ldapConfig
var log *multilog.MultiLogger

func init() {
	jsonHelper := helpers.Json{}
	if err := jsonHelper.UnmarshalJsonFile(LDAP_CONFIG_FILE, &ldapCfg); err != nil {
		panic(err)
	}
	if err := jsonHelper.UnmarshalJsonFile(ANOTHER_LDAP_CONFIG_FILE, &anotherLdapCfg); err != nil {
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
	app.Name = "remove-users-who-are-not-in-another-ldap"
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
	app.Usage = "remove users who are not in another LDAP."
	app.Action = func(c *cli.Context) {
		log.Info("ユーザ削除処理を開始します。")
		defer log.Info("ユーザ削除処理を終了しました。")

		action(c)
	}
	app.Run(os.Args)
}

func connectLdap(cfg ldapConfig) (*ldap.LDAPConnection, error) {
	l := ldap.NewLDAPConnection(cfg.Host, cfg.Port)
	if err := l.Connect(); err != nil {
		return nil, err
	}
	if cfg.BindDn != "" {
		if err := l.Bind(cfg.BindDn, cfg.BindPassword); err != nil {
			return nil, err
		}
	}
	if cfg.Verbose {
		l.Debug = true
	}
	return l, nil
}

func getUsersFromLdap(cfg ldapConfig) ([]ldapUser, error) {
	l, err := connectLdap(cfg)
	if err != nil {
		return nil, err
	}
	defer l.Close()

	searchRequest := ldap.NewSearchRequest(
		cfg.BaseDn,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		cfg.Filter,
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

func existsUser(users []ldapUser, uid string) bool {
	for _, user := range users {
		if uid == user.Uid {
			return true
		}
	}
	return false
}

func action(c *cli.Context) {
	isVerbose := c.Bool("verbose")
	if isVerbose {
		ldapCfg.Verbose = true
		anotherLdapCfg.Verbose = true
	}

	ldapUsers, err := getUsersFromLdap(ldapCfg)
	if err != nil {
		log.Error("LDAPサーバからのユーザ取得に失敗しました。" + err.Error())
		return
	}
	if len(ldapUsers) == 0 {
		log.Error("LDAPサーバにユーザが存在しません。" + LDAP_CONFIG_FILE + " を確認ください。")
		return
	}
	anotherLdapUsers, err := getUsersFromLdap(anotherLdapCfg)
	if err != nil {
		log.Error("別のLDAPサーバからのユーザ取得に失敗しました。" + err.Error())
		return
	}
	if len(anotherLdapUsers) == 0 {
		log.Error("別のLDAPサーバにユーザが存在しません。" + ANOTHER_LDAP_CONFIG_FILE + " を確認ください。")
		return
	}

	var hasError bool
	var userModel models.User
	existsUserCount := 0

	l, err := connectLdap(ldapCfg)
	if err != nil {
		log.Error("LDAP接続に失敗しました。" + err.Error())
		return
	}
	defer l.Close()

	for _, user := range ldapUsers {
		if existsUser(anotherLdapUsers, user.Uid) {
			existsUserCount++
			continue
		}
		if err := userModel.RemoveUser(l, user.Uid); err != nil {
			hasError = true
			log.Error("UserId: " + user.Uid + " ユーザの削除に失敗しました。" + err.Error())
		} else {
			log.Info("UserId: " + user.Uid + " ユーザを削除しました。")
		}
	}
	if hasError {
		log.Error("削除エラーになったものがあります。")
	}
	if existsUserCount == len(ldapUsers) {
		log.Info("削除対象のユーザはありませんでした。")
	}
}
