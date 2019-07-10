// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build darwin

package setting

import (
	"os"
	"path/filepath"

	"github.com/OpenPeeDeeP/xdg"
)

var (
	// fontPaths includes the paths for fonts on darwin.
	fontPaths = []string{
		filepath.Join(xdg.DataHome(), "Fonts"),
		filepath.Join(os.Getenv("HOME"), "Library", "Fonts"),
		"/Library/Fonts",
		"/Network/Library/Fonts",
		"/System/Library/Fonts",
		".",
	}
)
