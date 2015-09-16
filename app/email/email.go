package email

import (
	"app/config"
	"fmt"
	"io"
	netmail "net/mail"
	"os"
	"path"

	"github.com/aymerick/douceur/inliner"
	"github.com/hoisie/mustache"
	"github.com/jhillyerd/go.enmime"
	"golang.org/x/net/context"

	"google.golang.org/appengine/mail"
)

type Email mail.Message

const (
	ConfirmEmail = iota
	EmailReport  = iota
)

type IncomingEmail struct {
	Header netmail.Header
	Body   *enmime.MIMEBody
}

func Read(r io.Reader) (IncomingEmail, error) {
	var incomingMsg IncomingEmail
	msg, err := netmail.ReadMessage(r)
	if err != nil {
		return incomingMsg, err
	}
	incomingMsg = IncomingEmail{Header: msg.Header}
	incomingMsg.Body, err = enmime.ParseMIMEBody(msg)
	return incomingMsg, err
}

func New() *Email {
	return new(Email)
}

func (e *Email) LoadTemplate(id int, data map[string]interface{}) {
	var (
		templates = map[int][]string{
			0: []string{"confirm.email", "Email Confirmation"},
			1: []string{"report.email", "Your weekly Vine user report."},
			2: []string{"shareuser.email", "User Submission Confirmation"},
		}
		dir  = path.Join(os.Getenv("PWD"), "templates")
		body = mustache.RenderFile(path.Join(dir, templates[id][0]), data)
	)

	if id == 1 {
		link := fmt.Sprintf("https://www.davine.co/sign-up?op=Unsubscribe&id=%s", data["key"])
		e.Headers = map[string][]string{
			"List-Unsubscribe": {link},
		}
		var err error
		if e.HTMLBody, err = inliner.Inline(body); err != nil {
			e.HTMLBody = body
		}
	} else {
		e.Body = body
	}

	e.Subject = templates[id][1]
}

func (e Email) Send(c context.Context) error {
	e.Sender = config.Load(c)["emailSendAs"]
	msg := mail.Message(e)
	return mail.Send(c, &msg)
}
