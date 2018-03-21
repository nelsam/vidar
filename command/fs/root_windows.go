// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build windows

package fs

import (
	"fmt"
	"path/filepath"
	"strings"
)

func fsroot(path string) (string, bool) {
	if len(path) == 0 {
		// The drive letter hasn't been chosen
		return "", true
	}
	if i := strings.IndexRune(path, filepath.Separator); i == -1 || i == len(path)-1 {
		// The path is a drive root.
		driveEnd := strings.IndexRune(path, ':')
		if driveEnd == -1 {
			driveEnd = len(path)
		}
		drive := path[:driveEnd]
		return fmt.Sprintf(`%s:\`, drive), true
	}
	return "", false
}
