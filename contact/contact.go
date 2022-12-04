package contact

import (
	"fmt"
	"os"

	"github.com/vchimishuk/config"
)

var contactSpec = &config.Spec{
	Properties: []*config.PropertySpec{
		&config.PropertySpec{
			Type: config.TypeString,
			Name: "type",
		},
	},
}

type Contact interface {
	Name() string
	Notify(title string, body string) error
}

func ParseContact(file string) (Contact, error) {
	d, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c, err := config.Parse(contactSpec, string(d))
	if err != nil {
		return nil, err
	}

	t := c.String("type")
	switch t {
	case "email":
		return ParseEmail(file)
	default:
		return nil, fmt.Errorf("unsupported type %s", t)
	}
}
