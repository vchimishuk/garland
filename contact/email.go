package contact

import (
	"fmt"
	"net/smtp"
	"os"

	"github.com/vchimishuk/config"
	"github.com/vchimishuk/garland/fs"
)

var emailSpec = &config.Spec{
	Properties: []*config.PropertySpec{
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "from",
			Require: true,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "to",
			Require: true,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "host",
			Require: true,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "user",
			Require: true,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "pass",
			Require: true,
		},
	},
}

type Email struct {
	name string
	from string
	to   string
	host string
	user string
	pass string
}

func ParseEmail(file string) (*Email, error) {
	d, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c, err := config.Parse(emailSpec, string(d))
	if err != nil {
		return nil, err
	}

	return &Email{
		name: fs.Filename(file),
		from: c.String("from"),
		to:   c.String("to"),
		host: c.String("host"),
		user: c.String("user"),
		pass: c.String("pass"),
	}, nil
}

func (e *Email) Name() string {
	return e.name
}

func (e *Email) Notify(title string, body string) error {
	msg := fmt.Sprintf("To: %s\r\n"+
		"From: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n",
		e.to, e.from, title, body)
	auth := PlainAuth("", e.user, e.pass, e.host)

	return smtp.SendMail(e.host, auth, e.from, []string{e.to}, []byte(msg))
}
