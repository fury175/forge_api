package mailer

import (
	"embed"
)

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) (int, error)
}

//go:embed templates
var FS embed.FS

const (
	maxRetries = 3

	UserWelcomeTemplate = "user_invitation.tmpl"
)
