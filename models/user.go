package models

import (
	"fmt"
	"log"
	"time"

	"github.com/mavricknz/ldap"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
)

const (
	STATUS_NOT_INVITED = "0"
	STATUS_INVITED     = "1"
	STATUS_SIGN_UPED   = "2"
)

const LDAP_CONFIG_FILE = "config/ldap.json"

type ldapConfig struct {
	Host         string
	Port         uint16
	BindDn       string
	BindPassword string
	BaseDn       string
}

type InviteUser struct {
	Id         int64
	UserId     string `sql:"size:10"`
	LastName   string `sql:"size:16"`
	FirstName  string `sql:"size:16"`
	Mail       string `sql:"size:128"`
	InviteCode string `sql:"size:64"`
	Status     string `sql:"size:1"`
	InvitedAt  time.Time
	SignedUpAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

type User struct {
	UserId     string
	Password   string
	LastName   string
	FirstName  string
	Mail       string
	InviteCode string
}

var ldapCfg ldapConfig

func init() {
	jsonHelper := helpers.Json{}
	if err := jsonHelper.UnmarshalJsonFile(LDAP_CONFIG_FILE, &ldapCfg); err != nil {
		log.Fatalln(err)
	}
}

func (user *InviteUser) IsNotInvited() bool {
	return user.Status == STATUS_NOT_INVITED
}

func (user *InviteUser) IsSignUped() bool {
	return user.Status == STATUS_SIGN_UPED
}

func (user *User) addLdapUser(l *ldap.LDAPConnection) error {
	dn := "uid=" + user.UserId + "," + ldapCfg.BaseDn
	salt := fmt.Sprintf("%d%s", time.Now().UnixNano(), user.UserId)
	var addAttrs []ldap.EntryAttribute = []ldap.EntryAttribute{
		ldap.EntryAttribute{
			Name:   "objectclass",
			Values: []string{"person", "inetOrgPerson", "organizationalPerson", "top"},
		},
		ldap.EntryAttribute{
			Name:   "uid",
			Values: []string{user.UserId},
		},
		ldap.EntryAttribute{
			Name:   "cn",
			Values: []string{user.LastName + "　" + user.FirstName},
		},
		ldap.EntryAttribute{
			Name:   "userPassword",
			Values: []string{helpers.SSHA(user.Password, salt)},
		},
		ldap.EntryAttribute{
			Name:   "givenName",
			Values: []string{user.FirstName},
		},
		ldap.EntryAttribute{
			Name:   "sn",
			Values: []string{user.LastName},
		},
		ldap.EntryAttribute{
			Name:   "mail",
			Values: []string{user.Mail},
		},
	}
	addReq := ldap.NewAddRequest(dn)
	for _, attr := range addAttrs {
		addReq.AddAttribute(&attr)
	}
	fmt.Print(addReq)
	if err := l.Add(addReq); err != nil {
		return err
	}
	return nil
}

func (user *User) AddUser(ds *helpers.DataSource, inviteUser *InviteUser) error {
	l := ldap.NewLDAPConnection(ldapCfg.Host, ldapCfg.Port)
	if err := l.Connect(); err != nil {
		return err
	}
	defer l.Close()
	if err := l.Bind(ldapCfg.BindDn, ldapCfg.BindPassword); err != nil {
		return err
	}
	if exists, err := user.exists(l); err != nil {
		return err
	} else if exists {
		return AlreadyExistError{"ユーザーはすでに登録されています。"}
	}

	return ds.DoInTransaction(func(ds *helpers.DataSource) error {
		tx := ds.GetTx()
		now := time.Now()
		inviteUser.Status = STATUS_SIGN_UPED
		inviteUser.SignedUpAt = now
		inviteUser.UpdatedAt = now
		if err := tx.Save(inviteUser).Error; err != nil {
			return err
		}
		return user.addLdapUser(l)
	})
}

func (user *User) exists(l *ldap.LDAPConnection) (bool, error) {
	filter := "(uid=" + user.UserId + ")"
	attributes := []string{"uid"}
	searchRequest := ldap.NewSearchRequest(
		ldapCfg.BaseDn,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		filter,
		attributes,
		nil)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return false, err
	}
	if len(sr.Entries) == 0 {
		return false, nil
	}
	return true, nil
}

func (user *User) Validate() (bool, []string) {
	var messages []string
	if user.UserId == "" {
		messages = append(messages, "ユーザーIDを入力してください。")
	}
	if user.Password == "" {
		messages = append(messages, "パスワードを入力してください。")
	}
	if len(messages) > 0 {
		return false, messages
	}
	return true, messages
}

type AlreadyExistError struct {
	msg string
}

func (err AlreadyExistError) Error() string {
	return err.msg
}
