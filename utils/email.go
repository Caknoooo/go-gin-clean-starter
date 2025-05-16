package utils

import (
	"github.com/Caknoooo/go-gin-clean-starter/config"
	"gopkg.in/gomail.v2"
)

// dialer is an interface that matches gomail.Dialer for mocking purposes
type dialer interface {
	DialAndSend(...*gomail.Message) error
}

// newDialer is a variable that holds the function to create a new dialer
var newDialer = func(host string, port int, username, password string) dialer {
	return gomail.NewDialer(host, port, username, password)
}

func SendMail(toEmail string, subject string, body string) error {
	emailConfig, err := config.NewEmailConfig()
	if err != nil {
		return err
	}

	mailer := gomail.NewMessage()
	mailer.SetHeader("From", emailConfig.AuthEmail)
	mailer.SetHeader("To", toEmail)
	mailer.SetHeader("Subject", subject)
	mailer.SetBody("text/html", body)

	// Use the newDialer function instead of directly calling gomail.NewDialer
	dialer := newDialer(
		emailConfig.Host,
		emailConfig.Port,
		emailConfig.AuthEmail,
		emailConfig.AuthPassword,
	)

	err = dialer.DialAndSend(mailer)
	if err != nil {
		return err
	}

	return nil
}
