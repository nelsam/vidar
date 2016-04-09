// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build windows

package settings

import "os"

var (
	DefaultProject = Project{
		Name:   "*default*",
		Path:   os.Getenv("SYSTEMDRIVE"),
		Gopath: os.Getenv("GOPATH"),
	}
)
