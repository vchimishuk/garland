package source

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/vchimishuk/config"
	"github.com/vchimishuk/garland/fs"
)

var shellSpec = &config.Spec{
	Properties: []*config.PropertySpec{
		&config.PropertySpec{
			Type:    config.TypeString,
			Name:    "command",
			Require: true,
		},
	},
}

type Shell struct {
	name string
	cmd  string
}

func ParseShell(file string) (*Shell, error) {
	d, err := os.ReadFile(file)
	if err != nil {
		return nil, err
	}
	c, err := config.Parse(shellSpec, string(d))
	if err != nil {
		return nil, err
	}

	return &Shell{
		name: fs.Filename(file),
		cmd:  c.String("command"),
	}, nil
}

func (sh *Shell) Name() string {
	return sh.name
}

func (sh *Shell) Value() (float64, bool, error) {
	ctx, c := context.WithTimeout(context.Background(), 30*time.Minute)
	defer c()
	cmd := exec.CommandContext(ctx, "sh", "-c", sh.cmd)
	out, err := cmd.StdoutPipe()
	if err != nil {
		return 0, false, err
	}
	err = cmd.Start()
	if err != nil {
		return 0, false,
			fmt.Errorf("command `%s` failed: %w", sh.cmd, err)
	}
	b, err := io.ReadAll(out)
	if err != nil {
		return 0, false, err
	}
	err = cmd.Wait()
	if err != nil {
		return 0, false, fmt.Errorf("command `%s` failed: %w",
			sh.cmd, err)
	}

	s := strings.TrimSpace(string(b))
	if len(s) == 0 {
		return 0, true, nil
	}
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, false, fmt.Errorf("invalid integer value: %s", b)
	}

	return f, false, nil
}
