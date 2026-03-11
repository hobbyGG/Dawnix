package client

import (
	"context"
	"fmt"

	"gopkg.in/gomail.v2"
)

type EmailSender interface {
	Send(ctx context.Context, to, subject, body string) error
}

type emailClient struct {
	token string
	email string
	host  string
}

var _ EmailSender = (*emailClient)(nil)

func NewEmailClient(token, email string) EmailSender {
	return &emailClient{
		token: token,
		email: email,
		host:  "smtp.qq.com",
	}
}

func (ec *emailClient) Send(ctx context.Context, to, subject, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", ec.email)
	m.SetHeader("To", to)
	m.SetHeader("Subject", subject)
	m.SetBody("text/html", body)

	if ec.host == "" {
		return fmt.Errorf("failed to send email: invalid sender email %q", ec.email)
	}

	d := gomail.NewDialer(ec.host, 587, ec.email, ec.token)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}
	return nil
}
