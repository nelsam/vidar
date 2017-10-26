// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build windows

package focus

import "os"

var (
	userHome   = os.Getenv("USERPROFILE")
	systemRoot = os.Getenv("SYSTEMROOT")
)
