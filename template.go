package main

import (
	"bufio"
	"io"
	"os"
	"strings"

	"github.com/vchimishuk/garland/fs"
)

type Template struct {
	Name  string
	title string
	body  string
}

func ParseTemplate(file string) (*Template, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	r := bufio.NewReader(f)

	title, err := r.ReadString('\n')
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(r)
	if err != nil {
		return nil, err
	}

	return &Template{
		Name:  fs.Filename(file),
		title: strings.TrimSpace(title),
		body:  strings.TrimSpace(string(body)),
	}, nil
}

func (t *Template) Title(vars map[string]string) string {
	return t.eval(t.title, vars)
}

func (t *Template) Body(vars map[string]string) string {
	return t.eval(t.body, vars)
}

func (t *Template) eval(s string, vars map[string]string) string {
	for n, v := range vars {
		s = strings.ReplaceAll(s, "{"+n+"}", v)
	}

	return s
}
