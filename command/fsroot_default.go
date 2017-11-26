// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !windows

package command

func fsroot(path string) (string, bool) {
	switch path {
	case "", "/":
		return "/", true
	default:
		return "", false
	}
}
