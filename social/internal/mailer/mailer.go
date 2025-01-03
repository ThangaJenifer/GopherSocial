package mailer

import "embed"

const (
	FromName            = "GopherSocial"
	maxRetries          = 3
	UserWelcomeTemplate = "user_invitation.tmpl"
)

/*
https://gobyexample.com/embed-directive
checkout golang by Example where  allows programs to include arbitrary
files and folders in the Go binary at build time.
*/

//go:embed "templates"
var FS embed.FS

type Client interface {
	Send(templateFile, username, email string, data any, isSandbox bool) (int, error)
}
