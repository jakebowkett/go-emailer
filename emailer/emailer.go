package emailer

import (
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"net/smtp"
	"strings"
	"time"
)

type Emailer struct {
	Host    string
	Port    string
	User    string
	Pass    string
	From    string
	Name    string
	Timeout int
}

func (em *Emailer) Email(recipient, subject, body string) (err error) {

	addr := em.Host + ":" + em.Port
	auth := smtp.PlainAuth("", em.User, em.Pass, em.Host)

	email_parts := []string{
		fmt.Sprintf("To: %s", recipient),
		fmt.Sprintf(`From: "%s" <%s>`, em.Name, em.From),
		fmt.Sprintf("Subject: %s\r\n", subject),
		body,
	}
	msg := []byte(strings.Join(email_parts, "\r\n"))

	// Below is adapted from the standard library's
	// smtp.SendMail — we need it to have a timeout.
	timeout := time.Second * time.Duration(em.Timeout)
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return err
	}
	c, err := smtp.NewClient(conn, em.Host)
	if err != nil {
		return err
	}
	defer c.Close()
	if ok, _ := c.Extension("STARTTLS"); ok {
		config := &tls.Config{ServerName: em.Host}
		if err = c.StartTLS(config); err != nil {
			return err
		}
	}
	if ok, _ := c.Extension("AUTH"); !ok {
		return errors.New("server doesn't support AUTH")
	}
	if err := c.Auth(auth); err != nil {
		return err
	}
	if err := c.Mail(em.User); err != nil {
		return err
	}
	if err := c.Rcpt(recipient); err != nil {
		return err
	}

	// Issue DATA command to server. The writer returned
	// must be closed before calling any more methods on
	// the client.
	w, err := c.Data()
	if err != nil {
		return err
	}
	if _, err := w.Write(msg); err != nil {
		return err
	}
	if err := w.Close(); err != nil {
		return err
	}

	return c.Quit()
}
