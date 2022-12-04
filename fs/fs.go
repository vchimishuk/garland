package fs

import "path"

func Filename(file string) string {
	b := path.Base(file)
	e := path.Ext(b)

	return b[:len(b)-len(e)]
}
