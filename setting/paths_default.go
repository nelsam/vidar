// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !windows,!darwin

package setting

import (
	"os"
	"path/filepath"

	"github.com/OpenPeeDeeP/xdg"
)

var (
	// fontPaths includes the paths for fonts on default systems.
	// This should work for linux, BSD, and a few others.
	fontPaths = []string{
		filepath.Join(xdg.DataHome(), "fonts"),
		filepath.Join(os.Getenv("HOME"), ".fonts"),
		"/usr/local/share/fonts",
		"/usr/share/fonts",
		"/usr/X11/lib/X11/fonts",
	}
)
