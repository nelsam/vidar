// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !windows

package setting

import "os"

var (
	DefaultProject = Project{
		Name:   "*default*",
		Path:   "/",
		Gopath: os.Getenv("GOPATH"),
	}
)
