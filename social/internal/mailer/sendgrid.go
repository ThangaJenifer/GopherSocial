package mailer

import (
	"bytes"
	"fmt"
	"html/template"
	"time"

	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

// ex46
// All fields of this struct are private which is not exportable.
// so we will have a constructor function for this struct
type SendGridMailer struct {
	fromEmail string
	apiKey    string
	client    *sendgrid.Client
}

// Constructor function to initialize SendGridMailer
func NewSendGrid(apiKey, fromEmail string) *SendGridMailer {
	client := sendgrid.NewSendClient(apiKey)

	return &SendGridMailer{
		fromEmail: fromEmail,
		apiKey:    apiKey,
		client:    client,
	}

}

func (m *SendGridMailer) Send(templateFile, username, email string, data any, isSandbox bool) (int, error) {
	//FromName is email subject header e.g. whomever sending you email, it will be name of that person
	from := mail.NewEmail(FromName, m.fromEmail)
	to := mail.NewEmail(username, email)

	//template parsing and building
	tmpl, err := template.ParseFS(FS, "templates/"+templateFile)
	if err != nil {
		return -1, err
	}

	subject := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(subject, "subject", data)
	if err != nil {
		return -1, err
	}

	body := new(bytes.Buffer)
	err = tmpl.ExecuteTemplate(body, "body", data)
	if err != nil {
		return -1, err
	}

	//NewSingleEmail(from *mail.Email, subject string, to *mail.Email, plainTextContent string, htmlContent string) *mail.SGMailV3
	message := mail.NewSingleEmail(from, subject.String(), to, "", body.String())
	//Setting the sandbox value
	message.SetMailSettings(&mail.MailSettings{
		SandboxMode: &mail.Setting{
			Enable: &isSandbox,
		},
	})
	var retryErr error
	//sending message here
	for i := 0; i < maxRetries; i++ {
		response, retryErr := m.client.Send(message)
		if retryErr != nil {
			// Ex 47 These logs need to be handled by application struct which has sugared logger so see in auth.go file under registerUserhandler
			// log.Printf("failed to send email to %v, attempt %d of %d", email, i+1, maxRetries)
			// log.Printf("Error: %v", err.Error())

			//exponential backoff
			time.Sleep(time.Second * time.Duration(i+1))
			continue
		}

		//log.Printf("Email sent with status code %v", response.StatusCode)
		return response.StatusCode, nil
	}
	//last step when it fails to send email
	return -1, fmt.Errorf("failed to send email after %d attempts, error: %v", maxRetries, retryErr)

}
