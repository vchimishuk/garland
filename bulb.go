package main

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/vchimishuk/config"
	"github.com/vchimishuk/garland/fs"
	"github.com/vchimishuk/garland/source"
)

type Bulb struct {
	Name      string
	Interval  time.Duration
	Recovery  bool
	Source    source.Source
	Condition *Condition
	OnTpl     string
	OnVars    map[string]string
	OffTpl    string
	OffVars   map[string]string
}

var bulbSpec = &config.Spec{
	Properties: []*config.PropertySpec{
		&config.PropertySpec{
			Type: config.TypeDuration,
			Name: "interval",
		},
		&config.PropertySpec{
			Type:    config.TypeBool,
			Name:    "recovery",
			Require: false,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "source",
			Require: true,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "condition",
			Require: true,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "template.on",
			Require: true,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "template.on.*",
			Require: false,
			Repeat:  true,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "template.off",
			Require: false,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "template.off.*",
			Require: false,
			Repeat:  true,
		},
	},
}

func ParseBulb(file string) (*Bulb, error) {
	d, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c, err := config.Parse(bulbSpec, string(d))
	if err != nil {
		return nil, err
	}

	src, err := parseSource(c.String("source"), file)
	if err != nil {
		return nil, fmt.Errorf("failed to parse source: %w", err)
	}

	cn, err := ParseCondition(c.String("condition"))
	if err != nil {
		return nil, fmt.Errorf("invalid condition: %w", err)
	}

	return &Bulb{
		Name:      fs.Filename(file),
		Interval:  c.DurationOr("interval", time.Minute),
		Recovery:  c.BoolOr("recovery", true),
		Source:    src,
		Condition: cn,
		OnTpl:     c.String("template.on"),
		OnVars:    vars(c, "template.on."),
		OffTpl:    c.StringOr("template.off", ""),
		OffVars:   vars(c, "template.off."),
	}, nil
}

func parseSource(name string, file string) (source.Source, error) {
	switch name {
	case "graphite":
		return source.ParseGraphite(file)
	case "shell":
		return source.ParseShell(file)
	default:
		return nil, fmt.Errorf("unsupported source: %s", name)
	}
}

func vars(cfg *config.Config, prefix string) map[string]string {
	var v = make(map[string]string)
	for _, p := range cfg.Properties {
		if strings.HasPrefix(p.Name, prefix) {
			n := strings.TrimPrefix(p.Name, prefix)
			v[n] = cfg.String(p.Name)
		}
	}

	return v
}
