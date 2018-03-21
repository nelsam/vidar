// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build windows

package fs

import "os"

var (
	userHome = os.Getenv("USERPROFILE")

	// On Windows, we consider an empty string the system root
	// because the drive letter still needs to be chosen.  This
	// used to be os.Getenv("SYSTEMROOT"), but it prevented users
	// from opening files on other drives.
	systemRoot = ""
)
