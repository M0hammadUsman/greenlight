package mailer

import (
	"bytes"
	"embed"
	"gopkg.in/gomail.v2"
	"html/template"
)

/*
Below we declare a new variable with the type embed.FS (embedded file system) to hold our email templates.
This has a comment directive in the format `//go:embed <path>` IMMEDIATELY ABOVE it, which indicates to Go that we
want to store the contents of the ./templates directory in the templateFS embedded file system variable.
*/
//go:embed "templates"
var templateFS embed.FS

type Mailer struct {
	dialer *gomail.Dialer
	sender string
}

func New(host string, port int, username, password, sender string) Mailer {
	return Mailer{
		dialer: gomail.NewDialer(host, port, username, password),
		sender: sender,
	}
}

func (m Mailer) Send(recipient, templateFile string, data any) error {
	tmpl, err := template.New("email").ParseFS(templateFS, "templates/"+templateFile)
	if err != nil {
		return err
	}
	subject := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(subject, "subject", data); err != nil {
		return err
	}
	plainBody := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(plainBody, "plainBody", data); err != nil {
		return err
	}
	htmlBody := new(bytes.Buffer)
	if err = tmpl.ExecuteTemplate(htmlBody, "htmlBody", data); err != nil {
		return err
	}
	msg := gomail.NewMessage()
	msg.SetHeader("To", recipient)
	msg.SetHeader("From", m.sender)
	msg.SetHeader("Subject", subject.String())
	msg.SetBody("text/plain", plainBody.String())
	msg.AddAlternative("text/html", htmlBody.String())
	if err = m.dialer.DialAndSend(msg); err != nil {
		return err
	}
	return nil
}
