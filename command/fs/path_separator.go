// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !windows

package fs

import "path/filepath"

func pathSeparator(path string, r rune) bool {
	return r == filepath.Separator
}
