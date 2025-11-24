package mailers

import (
	"github.com/gobuffalo/buffalo/mail"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
)

// SendWelcomeEmails sends the welcome email to an address
func SendWelcomeEmails(to, name string) error {
	m := mail.NewMessage()
	email_url := envy.Get("STATEMENT_EMAIL", "")
	m.Subject = "Welcome"
	m.From = email_url
	m.To = []string{to}
	err := m.AddBody(r.HTML("welcome_email.html"), render.Data{
		"name": name,
	})
	if err != nil {
		return err
	}
	return smtp.Send(m)
}
