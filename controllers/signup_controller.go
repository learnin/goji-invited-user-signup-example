package controllers

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/flosch/pongo2"
	"github.com/mavricknz/ldap"
	"github.com/zenazn/goji/web"
)

type SignUpController struct {
}

type Form struct {
	HashKey         string
	Msg             string
	UserId          string
	Password        string
	ConfirmPassword string
	LastName        string
	FirstName       string
	Mail            string
}

var signupTpl = pongo2.Must(pongo2.FromFile("views/signup.tpl"))
var completeTpl = pongo2.Must(pongo2.FromFile("views/complete.tpl"))

func (controller *SignUpController) ShowSignupPage(c web.C, w http.ResponseWriter, r *http.Request) {
	hashKey := c.URLParams["hashKey"]
	if hashKey == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	var form Form
	form.HashKey = hashKey
	controller.renderSignupPage(c, w, r, form)
}

func (controller *SignUpController) ShowCompletePage(c web.C, w http.ResponseWriter, r *http.Request) {
	err := completeTpl.ExecuteWriter(pongo2.Context{"": ""}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (controller *SignUpController) renderSignupPage(c web.C, w http.ResponseWriter, r *http.Request, form Form) {
	err := signupTpl.ExecuteWriter(pongo2.Context{"form": form}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (controller *SignUpController) SignUp(c web.C, w http.ResponseWriter, r *http.Request) {
	var form Form
	form.UserId = r.FormValue("userId")
	form.Password = r.FormValue("password")
	form.ConfirmPassword = r.FormValue("confirmPassword")
	form.LastName = r.FormValue("lastName")
	form.FirstName = r.FormValue("firstName")
	form.Mail = r.FormValue("mail")
	form.HashKey = r.FormValue("hashKey")
	if form.HashKey == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	if !controller.validate(&form) {
		form.Msg = strings.Replace(form.Msg, "\n", "<br/>", -1)
		controller.renderSignupPage(c, w, r, form)
		return
	}
	fmt.Println(controller.hash(form.UserId))
	if controller.hash(form.UserId) != form.HashKey {
		form.Msg = "ユーザーIDを正しく入力してください。"
		controller.renderSignupPage(c, w, r, form)
		return
	}
	alreadyExist, err := controller.existsUser()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if alreadyExist {
		form.Msg = "そのユーザーはすでに登録されています。"
		controller.renderSignupPage(c, w, r, form)
		return
	}
	http.Redirect(w, r, "/signup/complete", http.StatusFound)
}

func (controller *SignUpController) addUser(l *ldap.LDAPConnection) error {
	dn := "cn=test3,cn=users,cn=accounts,dc=demo1,dc=freeipa,dc=org"
	var addAttrs []ldap.EntryAttribute = []ldap.EntryAttribute{
		ldap.EntryAttribute{
			Name: "objectclass",
			Values: []string{
				"person", "inetOrgPerson", "organizationalPerson", "top",
			},
		},
		ldap.EntryAttribute{
			Name: "uid",
			Values: []string{
				"test3",
			},
		},
		ldap.EntryAttribute{
			Name: "cn",
			Values: []string{
				"test3",
			},
		},
		ldap.EntryAttribute{
			Name: "givenName",
			Values: []string{
				"test3gn",
			},
		},
		ldap.EntryAttribute{
			Name: "sn",
			Values: []string{
				"test3sn",
			},
		},
	}
	addReq := ldap.NewAddRequest(dn)
	for _, attr := range addAttrs {
		addReq.AddAttribute(&attr)
	}
	fmt.Print(addReq)
	err := l.Add(addReq)
	if err != nil {
		return err
	}
	return nil
}

func (controller *SignUpController) existsUser() (bool, error) {
	ldapServer := "ipa.demo1.freeipa.org"
	l := ldap.NewLDAPConnection(ldapServer, 389)
	err := l.Connect()
	if err != nil {
		return false, err
	}
	defer l.Close()
	err = l.Bind("uid=admin,cn=users,cn=accounts,dc=demo1,dc=freeipa,dc=org", "Secret123")
	if err != nil {
		return false, err
	}
	baseDN := "cn=users,cn=accounts,dc=demo1,dc=freeipa,dc=org"
	var filter []string = []string{"(uid=test3)"}
	var attributes []string = []string{
		"uid", "givenname"}
	searchRequest := ldap.NewSearchRequest(
		baseDN,
		ldap.ScopeWholeSubtree, ldap.DerefAlways, 0, 0, false,
		filter[0],
		attributes,
		nil)
	sr, err := l.Search(searchRequest)
	if err != nil {
		return false, err
	}
	if len(sr.Entries) == 0 {
		return false, controller.addUser(l)
	}
	return true, nil
}

func (controller *SignUpController) validate(form *Form) bool {
	if form.UserId == "" {
		form.Msg += "ユーザーIDを入力してください。\n"
	}
	if form.Password == "" {
		form.Msg += "パスワードを入力してください。\n"
	}
	if form.ConfirmPassword == "" {
		form.Msg += "パスワード(確認)を入力してください。\n"
	}
	if form.LastName == "" {
		form.Msg += "姓を入力してください。\n"
	}
	if form.FirstName == "" {
		form.Msg += "名を入力してください。\n"
	}
	if form.Mail == "" {
		form.Msg += "メールアドレスを入力してください。\n"
	}
	if form.Password != form.ConfirmPassword {
		form.Msg += "パスワードとパスワード(確認)が一致していません。\n"
	}
	if form.Msg != "" {
		return false
	}
	return true
}

func (controller *SignUpController) hash(s string) string {
	const salt = "HsE@U91Ie!8ye8ay^e87wya7Y*R%38[0(*T[9w4eut[9e"
	hash := sha256.New()
	io.WriteString(hash, s+salt)
	return hex.EncodeToString(hash.Sum(nil))
}
