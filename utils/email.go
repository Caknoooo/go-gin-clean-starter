package utils

import (
	"github.com/Caknoooo/go-gin-clean-starter/config"
	"gopkg.in/gomail.v2"
)

// Dialer is an interface that matches gomail.Dialer for mocking purposes
type Dialer interface {
	DialAndSend(...*gomail.Message) error
}

// NewDialer is a variable that holds the function to create a new Dialer
var NewDialer = func(host string, port int, username, password string) Dialer {
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

	// Use the NewDialer function instead of directly calling gomail.NewDialer
	dialer := NewDialer(
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
