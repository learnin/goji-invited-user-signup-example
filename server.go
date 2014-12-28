package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"

	"github.com/flosch/pongo2"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
)

type Page struct {
	HashKey string
	Msg     string
	UserId  string
}

var indexTpl = pongo2.Must(pongo2.FromFile("views/index.tpl"))
var completeTpl = pongo2.Must(pongo2.FromFile("views/complete.tpl"))

func showIndexPage(c web.C, w http.ResponseWriter, r *http.Request) {
	hashKey := c.URLParams["hashKey"]
	if hashKey == "" {
		http.Error(w, "", http.StatusNotFound)
		return
	}
	var page Page
	page.HashKey = hashKey
	renderIndex(c, w, r, page)
}

func showCompletePage(c web.C, w http.ResponseWriter, r *http.Request) {
	err := completeTpl.ExecuteWriter(pongo2.Context{"": ""}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func renderIndex(c web.C, w http.ResponseWriter, r *http.Request, page Page) {
	err := indexTpl.ExecuteWriter(pongo2.Context{"page": page}, w)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func authentication(c web.C, w http.ResponseWriter, r *http.Request) {
	var page Page
	page.UserId = r.FormValue("userId")
	page.HashKey = r.FormValue("hashKey")
	if page.UserId == "" {
		page.Msg = "ユーザーIDを入力してください。"
		renderIndex(c, w, r, page)
		return
	}
	fmt.Println(hash(page.UserId))
	if hash(page.UserId) != page.HashKey {
		page.Msg = "ユーザーIDを正しく入力してください。"
		renderIndex(c, w, r, page)
		return
	}
	http.Redirect(w, r, "/complete", http.StatusFound)

}

func hash(s string) string {
	const salt = "HsE@U91Ie!8ye8ay^e87wya7Y*R%38[0(*T[9w4eut[9e"
	hash := sha256.New()
	io.WriteString(hash, s+salt)
	return hex.EncodeToString(hash.Sum(nil))
}

func main() {
	goji.Get("/:hashKey", showIndexPage)
	goji.Post("/authentication", authentication)
	goji.Get("/complete", showCompletePage)
	goji.Serve()
}
