// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build windows

package fs

import "path/filepath"

func pathSeparator(path string, r rune) bool {
	switch r {
	case ':':
		for _, r := range path {
			if r == filepath.Separator {
				return false
			}
		}
		fallthrough
	case filepath.Separator:
		// TODO: should we include '/' as well?
		// windows seems to be happy to allow '/' as a path separator.
		return true
	}
	return false
}
