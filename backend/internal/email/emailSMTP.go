package email

import (
	"fmt"
	"net/smtp"
)

type emailSMTP struct {
	password string
	username string
	smtpHost string
}

func NewEmailSMTP(host, user, pass string) *emailSMTP {
	return &emailSMTP{smtpHost: host, username: user, password: pass}
}

func (e *emailSMTP) SendNotification(to, body string) error {

	message := []byte("Subject: You have como.peculiarity notifications!\r\n\r\n" + body + "\r\n")
	auth := smtp.PlainAuth("", e.username, e.password, e.smtpHost)
	err := smtp.SendMail(e.smtpHost+":587", auth, e.username, []string{to}, message)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Email to:" + to + " sent Successfully")
	return nil
}
