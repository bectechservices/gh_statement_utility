package mailers

import (
	"github.com/gobuffalo/buffalo/mail"
	"github.com/gobuffalo/buffalo/render"
	"github.com/gobuffalo/envy"
)

// SendPasswordResets sends the password reset mail
func SendPasswordResets(to, name, url string) error {
	m := mail.NewMessage()
	email_url := envy.Get("STATEMENT_EMAIL", "")
	// fill in with your stuff:
	m.Subject = "Password Reset"
	m.From = email_url
	m.To = []string{to}
	err := m.AddBody(r.HTML("password_reset.html"), render.Data{
		"name": name,
		"link": url,
	})

	if err != nil {
		return err
	}
	return smtp.Send(m)
}
