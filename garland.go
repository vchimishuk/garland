package main

import (
	"fmt"
	"log"
	"os"
	"path"
	"strings"
	"time"

	"github.com/vchimishuk/garland/contact"
	"github.com/vchimishuk/garland/slices"
	"golang.org/x/exp/maps"
)

type Notice struct {
	Bulb  *Bulb
	State bool
	Value float64
}

type DeferedNotice struct {
	Time    time.Time
	Delay   time.Duration
	Notice  *Notice
	Contact contact.Contact
}

const ConfigDir string = "/etc/garland"

var Notifications = make(chan *Notice, 10)

func state(s bool) string {
	if s {
		return "on"
	} else {
		return "off"
	}
}

func listDir(dir string, ext string) ([]string, error) {
	es, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range es {
		if e.IsDir() {
			continue
		}

		if !strings.HasSuffix(e.Name(), "."+ext) {
			continue
		}

		files = append(files, e.Name())
	}

	return files, nil
}

func monitor(b *Bulb) {
	t := time.NewTicker(b.Interval)

	st := false
	nerr := 0
	for {
		select {
		case _ = <-t.C:
			log.Printf("Checking %s.", b.Name)
			v, n, err := b.Source.Value()
			if err != nil {
				log.Printf("Failed to get %s's value: %s",
					b.Source.Name(), err)
				t.Reset(time.Minute)
				nerr++
			} else {
				if n {
					log.Printf("%s's value is null.",
						b.Name)
				} else {
					log.Printf("%s's value is %f.",
						b.Name, v)
				}
				s := b.Condition.Eval(v, n)
				if s != st {
					log.Printf("%s goes %s.",
						b.Name, state(s))
					Notifications <- &Notice{
						Bulb:  b,
						State: s,
						Value: v,
					}
				}
				st = s

				if nerr != 0 {
					t.Reset(b.Interval)
					nerr = 0
				}
			}
		}
	}
}

func parseContacts() ([]contact.Contact, error) {
	dir := path.Join(ConfigDir, "contacts")
	files, err := listDir(dir, "conf")
	if err != nil {
		return nil,
			fmt.Errorf("failed to parse %s folder: %w",
				dir, err)
	}
	var cs []contact.Contact
	for _, f := range files {
		cf := path.Join(dir, f)
		c, err := contact.ParseContact(cf)
		if err != nil {
			return nil, fmt.Errorf("failed to parse %s: %w",
				cf, err)
		}
		cs = append(cs, c)
	}

	return cs, nil
}

func findTpl(tpls []*Template, name string) *Template {
	for _, t := range tpls {
		if t.Name == name {
			return t
		}
	}

	return nil
}

func parseTemplates() ([]*Template, error) {
	dir := path.Join(ConfigDir, "templates")
	files, err := listDir(dir, "tpl")
	if err != nil {
		return nil,
			fmt.Errorf("failed to parse %s folder: %w",
				dir, err)
	}

	var tpls []*Template
	for _, f := range files {
		tf := path.Join(dir, f)
		t, err := ParseTemplate(tf)
		if err != nil {
			return nil,
				fmt.Errorf("failed to parse %s: %w", tf, err)
		}
		tpls = append(tpls, t)
	}

	return tpls, nil
}

func parseBulbs(tpls []*Template) ([]*Bulb, error) {
	dir := path.Join(ConfigDir, "bulbs")
	files, err := listDir(dir, "conf")
	if err != nil {
		return nil,
			fmt.Errorf("failed to parse %s folder: %w",
				dir, err)
	}
	var bulbs []*Bulb
	for _, f := range files {
		bf := path.Join(dir, f)
		b, err := ParseBulb(bf)
		if err != nil {
			return nil,
				fmt.Errorf("failed to parse %s: %w", bf, err)
		}
		if (b.OnTpl != "" && findTpl(tpls, b.OnTpl) == nil) ||
			(b.OffTpl != "" && findTpl(tpls, b.OffTpl) == nil) {
			return nil,
				fmt.Errorf("failed to parse %s: "+
					"template %s not found",
					bf, b.OnTpl)
		}

		bulbs = append(bulbs, b)
	}

	return bulbs, nil
}

func incDeferTime(d time.Duration) time.Duration {
	if d == 0 {
		return time.Minute
	}

	n := d * 2
	if n.Hours() > 1 {
		n = time.Hour
	}

	return n
}

func notify(n *Notice, c contact.Contact, tpls []*Template) (bool, error) {
	tplIdx := slices.ToMap(tpls, func(t *Template) string {
		return t.Name
	})

	var tpl *Template
	var vars map[string]string

	vars = map[string]string{
		"bulb":   n.Bulb.Name,
		"status": state(n.State),
		"time":   time.Now().Format(time.RFC822),
		"value":  fmt.Sprintf("%f", n.Value),
	}
	if n.State && n.Bulb.OnTpl != "" {
		tpl = tplIdx[n.Bulb.OnTpl]
		maps.Copy(vars, n.Bulb.OnVars)
	} else if n.Bulb.OffTpl != "" {
		tpl = tplIdx[n.Bulb.OffTpl]
		maps.Copy(vars, n.Bulb.OffVars)
	}

	if tpl != nil {
		title := tpl.Title(vars)
		body := tpl.Body(vars)
		err := c.Notify(title, body)
		if err != nil {
			return false, err
		}
	}

	return tpl != nil, nil
}

func notifier(contacts []contact.Contact, tpls []*Template) {
	t := time.NewTicker(time.Minute)
	var defered []*DeferedNotice

	for {
		var n *Notice

		select {
		case _ = <-t.C:
		case n = <-Notifications:
			for _, c := range contacts {
				d := &DeferedNotice{
					Time:    time.Now(),
					Delay:   0,
					Notice:  n,
					Contact: c,
				}
				defered = append([]*DeferedNotice{d},
					defered...)
			}
		}

		for i := 0; i < len(defered) && len(defered) > 0; {
			d := defered[1]
			if d.Time.Add(d.Delay).After(time.Now()) {
				i++
				continue
			}

			defered = defered[1:]
			sent, err := notify(d.Notice, d.Contact, tpls)
			if err != nil {
				d.Delay = incDeferTime(d.Delay)
				i++
				log.Printf("Failed to notify %s "+
					"for went %s %s: %s. "+
					"Will retry later.",
					d.Contact.Name(), state(d.Notice.State),
					d.Notice.Bulb.Name, err)

			} else {
				defered = append(defered[:i], defered[i+1:]...)
				if sent {
					log.Printf("Notified %s for went %s %s.",
						d.Contact.Name(),
						state(d.Notice.State),
						d.Notice.Bulb.Name)
				}
			}
		}
	}
}

func main() {
	log.Println("Starting Garland...")

	contacts, err := parseContacts()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded contacts: %s", strings.Join(slices.Map(contacts,
		func(c contact.Contact) string { return c.Name() }), ", "))

	tpls, err := parseTemplates()
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Loaded templates: %s", strings.Join(slices.Map(tpls,
		func(t *Template) string { return t.Name }), ", "))

	go notifier(contacts, tpls)

	bulbs, err := parseBulbs(tpls)
	if err != nil {
		log.Fatal(err)
	}
	for _, b := range bulbs {
		go monitor(b)
	}
	log.Printf("Loaded bulbs: %s", strings.Join(slices.Map(bulbs,
		func(b *Bulb) string { return b.Name }), ", "))

	// Sleep forever.
	for {
		time.Sleep(time.Hour)
	}
}
