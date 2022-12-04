package source

import (
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/vchimishuk/config"
	"github.com/vchimishuk/garland/fs"
)

var graphiteSpec = &config.Spec{
	Properties: []*config.PropertySpec{
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "host",
			Require: true,
		},
		&config.PropertySpec{
			Type:    config.TypeDuration,
			Name:    "period",
			Require: true,
		},
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "query",
			Require: true,
		},
	},
}

type Graphite struct {
	name   string
	host   string
	period time.Duration
	query  string
	client *http.Client
}

func ParseGraphite(file string) (*Graphite, error) {
	d, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c, err := config.Parse(graphiteSpec, string(d))
	if err != nil {
		return nil, err
	}

	return &Graphite{
		name:   fs.Filename(file),
		host:   c.String("host"),
		period: c.Duration("period"),
		query:  c.String("query"),
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}, nil
}

func (g *Graphite) Name() string {
	return g.name
}

func (g *Graphite) Value() (float64, bool, error) {
	uri := fmt.Sprintf("http://%s/render?format=csv&from=-%.0fs&target=%s",
		g.host, g.period.Seconds(), url.QueryEscape(g.query))
	resp, err := g.client.Get(uri)
	if err != nil {
		return 0, false, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 0, false, fmt.Errorf("non-ok (%d) server response",
			resp.StatusCode)
	}

	r := bufio.NewReader(resp.Body)
	l := ""
	for {
		s, err := r.ReadString('\n')
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return 0, false, err
		}
		l = s
	}
	if l == "" {
		return 0, false, errors.New("invalid server response")
	}

	cr := csv.NewReader(strings.NewReader(l))
	rec, err := cr.Read()
	if err != nil {
		return 0, false, err
	}
	if len(rec) == 0 {
		return 0, false, errors.New("invalid server reponse")
	}
	v := rec[len(rec)-1]
	if v == "" {
		return 0, true, nil
	}
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return 0, false, fmt.Errorf("invalid value: %s", v)
	}

	return f, false, nil
}
