package mailer

import (
	"bytes"
	"fmt"
	"text/template"
	"time"

	"gopkg.in/gomail.v2"
)

type mailTrapClient struct {
	fromEmail       string
	apiKey          string
	sandboxUsername string
}

func NewMailTrapClient(fromEmail string, apiKey string, sandboxUsername string) (Client, error) {
	if apiKey == "" {
		return mailTrapClient{}, fmt.Errorf("Mailtrap api key not set")
	}

	return mailTrapClient{
		fromEmail:       fromEmail,
		apiKey:          apiKey,
		sandboxUsername: sandboxUsername,
	}, nil
}

func (m mailTrapClient) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	if err := tmpl.ExecuteTemplate(body, "body", data); err != nil {
		return -1, err
	}

	message := gomail.NewMessage()
	message.SetHeader("From", m.fromEmail)
	message.SetHeader("To", email)
	message.SetHeader("Subject", subject.String())
	message.AddAlternative("text/html", body.String())

	var dialer *gomail.Dialer
	if isSandbox {
		dialer = gomail.NewDialer("sandbox.smtp.mailtrap.io", 2525, m.sandboxUsername, m.apiKey)
	} else {
		dialer = gomail.NewDialer("live.smtp.mailtrap.io", 587, "api", m.apiKey)
	}

	var retryErr error
	for i := range maxRetries {
		retryErr = dialer.DialAndSend(message)
		if retryErr != nil {
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}
		break
	}

	if retryErr != nil {
		return -1, fmt.Errorf("failed to send email after %d attempt, error: %w", maxRetries, retryErr)
	}

	return 200, nil
}
