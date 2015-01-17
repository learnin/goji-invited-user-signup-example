package helpers

import (
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

type SmtpClient struct {
	Host     string
	Port     uint16
	Username string
	Password string
}

type Mail struct {
	From    string
	To      string
	Subject string
	Body    string
}

func (client *SmtpClient) Connect() (*smtp.Client, error) {
	c, err := smtp.Dial(fmt.Sprintf("%s:%d", client.Host, client.Port))
	if err != nil {
		return nil, err
	}
	if err := c.Hello("localhost"); err != nil {
		return nil, err
	}
	if client.Username != "" {
		if ok, _ := c.Extension("AUTH"); ok {
			if err := c.Auth(smtp.PlainAuth("", client.Username, client.Password, client.Host)); err != nil {
				return nil, err
			}
		}
	}
	return c, nil
}

func (client *SmtpClient) SendMail(c *smtp.Client, mail Mail) error {
	if err := c.Reset(); err != nil {
		return err
	}
	c.Mail(mail.From)
	if err := c.Rcpt(mail.To); err != nil {
		return err
	}
	w, err := c.Data()
	if err != nil {
		return err
	}
	subject, err := encodeSubject(mail.Subject)
	if err != nil {
		return err
	}
	body, err := encodeToJIS(mail.Body + "\r\n")
	if err != nil {
		return err
	}
	msg := "From: " + mail.From + "\r\n" +
		"To: " + mail.To + "\r\n" +
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
