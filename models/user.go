package models

import (
	"fmt"

	"github.com/mavricknz/ldap"

	"github.com/learnin/goji-invited-user-signup-example/helpers"
)

const SALT = "HsE@U91Ie!8ye8ay^e87wya7Y*R%38[0(*T[9w4eut[9e"
const LDAP_SERVER = "ipa.demo1.freeipa.org"
const LDAP_PORT = 389
const LDAP_BIND_USER = "uid=admin,cn=users,cn=accounts,dc=demo1,dc=freeipa,dc=org"
const LDAP_BIND_PASSWORD = "Secret123"
const LDAP_BASE_DN = "cn=users,cn=accounts,dc=demo1,dc=freeipa,dc=org"

type User struct {
	UserId    string
	Password  string
	LastName  string
	FirstName string
	Mail      string
	HashKey   string
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

func (user *User) AddUser() error {
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
	return user.addLdapUser(l)
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

func (user *User) ValidateHashKey() bool {
	fmt.Println(helpers.Hash(user.UserId, SALT))
	return helpers.Hash(user.UserId, SALT) == user.HashKey
}

type AlreadyExistError struct {
	msg string
}

func (err AlreadyExistError) Error() string {
	return err.msg
}
