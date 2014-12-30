package models

import (
	"fmt"
	"time"

	"github.com/mavricknz/ldap"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
)

const LDAP_SERVER = "ipa.demo1.freeipa.org"
const LDAP_PORT = 389
const LDAP_BIND_USER = "uid=admin,cn=users,cn=accounts,dc=demo1,dc=freeipa,dc=org"
const LDAP_BIND_PASSWORD = "Secret123"
const LDAP_BASE_DN = "cn=users,cn=accounts,dc=demo1,dc=freeipa,dc=org"
const (
	STATUS_NOT_INVITED = "0"
	STATUS_INVITED     = "1"
	STATUS_SIGN_UPED   = "2"
)

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
	UserId    string
	Password  string
	LastName  string
	FirstName string
	Mail      string
	HashKey   string
}

func (user *InviteUser) IsSignUped() bool {
	return user.Status == STATUS_SIGN_UPED
}

func (user *User) addLdapUser(l *ldap.LDAPConnection) error {
	dn := "uid=" + user.UserId + ",cn=users,cn=accounts,dc=demo1,dc=freeipa,dc=org"
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
	l := ldap.NewLDAPConnection(LDAP_SERVER, LDAP_PORT)
	if err := l.Connect(); err != nil {
		return err
	}
	defer l.Close()
	if err := l.Bind(LDAP_BIND_USER, LDAP_BIND_PASSWORD); err != nil {
		return err
	}
	if exists, err := user.exists(l); err != nil {
		return err
	} else if exists {
		return AlreadyExistError{"ユーザーはすでに登録されています。"}
	}

	f := func(ds *helpers.DataSource) error {
		tx := ds.GetTx()
		inviteUser.Status = STATUS_SIGN_UPED
		inviteUser.SignedUpAt = time.Now()
		if err := tx.Save(inviteUser).Error; err != nil {
			return err
		}
		return user.addLdapUser(l)
	}
	if err := ds.DoInTransaction(f); err != nil {
		return err
	}
	return nil
}

func (user *User) exists(l *ldap.LDAPConnection) (bool, error) {
	filter := "(uid=" + user.UserId + ")"
	attributes := []string{"uid"}
	searchRequest := ldap.NewSearchRequest(
		LDAP_BASE_DN,
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

func (user *User) Validate() (bool, string) {
	var msg string
	if user.UserId == "" {
		msg += "ユーザーIDを入力してください。\n"
	}
	if user.Password == "" {
		msg += "パスワードを入力してください。\n"
	}
	if user.LastName == "" {
		msg += "姓を入力してください。\n"
	}
	if user.FirstName == "" {
		msg += "名を入力してください。\n"
	}
	if user.Mail == "" {
		msg += "メールアドレスを入力してください。\n"
	}
	if msg != "" {
		return false, msg
	}
	return true, ""
}

type AlreadyExistError struct {
	msg string
}

func (err AlreadyExistError) Error() string {
	return err.msg
}
