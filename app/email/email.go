package email

import (
	"app/config"
	"appengine"
	"appengine/mail"
	"github.com/hoisie/mustache"
	"os"
	"path"
)

type Email mail.Message

const (
	ConfirmEmail = iota
	EmailReport  = iota
)

func New() *Email {
	return new(Email)
}

func (e *Email) LoadTemplate(id int, data map[string]string) {
	var (
		templates = map[int][]string{
			0: []string{"confirm.email", "Email Confirmation"},
			1: []string{"report.email", "Your weekly Vine user report."},
		}
		dir = path.Join(os.Getenv("PWD"), "templates")
	)
	if id == 1 {
		e.Headers = map[string][]string{
			"List-Unsubscribe": {data["unsubscribe"]},
		}
	}
	e.Subject = templates[id][1]
	e.Body = mustache.RenderFile(path.Join(dir, templates[id][0]), data)
}

func (e Email) Send(c appengine.Context) error {
	e.Sender = config.Load(c)["emailSendAs"]
	msg := mail.Message(e)
	return mail.Send(c, &msg)
}
