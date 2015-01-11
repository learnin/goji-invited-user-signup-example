package helpers

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"log"
	"net/smtp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

const SMTP_CONFIG_FILE = "config/smtp.json"

type SmtpUtil struct {
}

type smtpConfig struct {
	Host     string
	Port     uint16
	Username string
	Password string
	From     string
}

var smtpCfg smtpConfig

func init() {
	jsonHelper := Json{}
	if err := jsonHelper.UnmarshalJsonFile(SMTP_CONFIG_FILE, &smtpCfg); err != nil {
		log.Fatalln(err)
	}
}

func (util *SmtpUtil) Connect() (*smtp.Client, error) {
	c, err := smtp.Dial(fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port))
	if err != nil {
		return nil, err
	}

	// auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)

	if err := c.Hello("localhost"); err != nil {
		return nil, err
	}
	// if ok, _ := c.Extension("AUTH"); ok {
	// 	if err := c.Auth(auth); err != nil {
	// 		return err
	// 	}
	// }
	return c, nil
}

func (util *SmtpUtil) SendMail(c *smtp.Client, to string) error {
	if err := c.Reset(); err != nil {
		return err
	}
	c.Mail(smtpCfg.From)
	if err := c.Rcpt(to); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	subject, err := encodeSubject("テストおおおおおお。あああいいいいんんんaいいう1234あああああああああああいいいいいいいいいいう")
	if err != nil {
		return err
	}
	body, err := encodeToJIS("テストメールです" + "\r\n")
	if err != nil {
		return err
	}
	msg := "From: " + smtpCfg.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject:" + subject +
		"Date: " + time.Now().Format(time.RFC1123Z) + "\r\n" +
		"MIME-Version: 1.0\r\n" +
		"Content-Type: text/plain; charset=ISO-2022-JP\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"\r\n" +
		body
	if _, err = w.Write([]byte(msg)); err != nil {
		return err
	}
	return w.Close()
}

func encodeToJIS(s string) (string, error) {
	r, err := ioutil.ReadAll(transform.NewReader(strings.NewReader(s), japanese.ISO2022JP.NewEncoder()))
	if err != nil {
		return "", err
	}
	return string(r), nil
}

func encodeSubject(subject string) (string, error) {
	b := make([]byte, 0, utf8.RuneCountInString(subject))
	for _, s := range splitByCharLength(subject, 13) {
		b = append(b, " =?ISO-2022-JP?B?"...)
		s, err := encodeToJIS(s)
		if err != nil {
			return "", err
		}
		b = append(b, base64.StdEncoding.EncodeToString([]byte(s))...)
		b = append(b, "?=\r\n"...)
	}
	return string(b), nil
}

func splitByCharLength(s string, length int) []string {
	result := []string{}
	b := make([]byte, 0, length)
	for i, c := range strings.Split(s, "") {
		b = append(b, c...)
		if i%length == 0 {
			result = append(result, string(b))
			b = make([]byte, 0, length)
		}
	}
	if len(b) > 0 {
		result = append(result, string(b))
	}
	return result
}

// FIXME 接続は1回にして、Reset, Mail, Rcpt, Data, w.Closeを宛先文回す。
// func sendMail(inviteUser models.InviteUser) error {
// 	c, err := smtp.Dial(fmt.Sprintf("%s:%d", smtpCfg.Host, smtpCfg.Port))
// 	if err != nil {
// 		return err
// 	}
// 	defer c.Close()

// 	// auth := smtp.PlainAuth("", smtpCfg.Username, smtpCfg.Password, smtpCfg.Host)

// 	if err := c.Hello("localhost"); err != nil {
// 		return err
// 	}
// 	// if ok, _ := c.Extension("AUTH"); ok {
// 	// 	if err := c.Auth(auth); err != nil {
// 	// 		return err
// 	// 	}
// 	// }

// 	c.Mail(smtpCfg.From)
// 	if err := c.Rcpt(inviteUser.Mail); err != nil {
// 		return err
// 	}
// 	w, err := c.Data()
// 	if err != nil {
// 		return err
// 	}
// 	// FIXME エンコード
// 	msg := "Content-Type: text/plain; charset=ISO-2022-JP\r\n" +
// 		"Content-Transfer-Encoding: 7bit\r\n" +
// 		"From: " + smtpCfg.From + "\r\n" +
// 		"To: " + inviteUser.Mail + "\r\n" +
// 		"Subject: test\r\n" +
// 		"\r\n" +
// 		"テストメールです"
// 	if _, err = w.Write([]byte(msg)); err != nil {
// 		return err
// 	}
// 	if err := w.Close(); err != nil {
// 		return err
// 	}
// 	return c.Quit()
// }
