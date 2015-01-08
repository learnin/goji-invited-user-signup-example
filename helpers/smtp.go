package helpers

import (
	"fmt"
	"log"
	"net/smtp"
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
	// FIXME エンコード
	msg := "Content-Type: text/plain; charset=ISO-2022-JP\r\n" +
		"Content-Transfer-Encoding: 7bit\r\n" +
		"From: " + smtpCfg.From + "\r\n" +
		"To: " + to + "\r\n" +
		"Subject: test\r\n" +
		"\r\n" +
		"テストメールです"
	if _, err = w.Write([]byte(msg)); err != nil {
		return err
	}
	return w.Close()
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
