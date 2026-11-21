package mailer

import "embed"

//go:embed templates
var FS embed.FS

const (
	maxRetries = 3

	UserWelcomeTemplate = "user_invitation.tmpl"
)
