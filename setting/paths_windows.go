// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build windows

package setting

import (
	"os"
	"path/filepath"

	"github.com/OpenPeeDeeP/xdg"
)

var (
	fontPaths = []string{
		filepath.Join(xdg.DataHome(), "FONTS"),
		filepath.Join(os.Getenv("SYSTEMDRIVE"), "WINDOWS", "FONTS"),
		filepath.Join(os.Getenv("SYSTEMDRIVE")+"//", "WINDOWS", "FONTS"),
	}
)
