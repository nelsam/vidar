// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package fs

import (
	"log"
	"path/filepath"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
)

type dirLabel struct {
	mixins.Label

	driver gxui.Driver
}

func newDirLabel(driver gxui.Driver, theme *basic.Theme) *dirLabel {
	label := &dirLabel{driver: driver}
	label.Label.Init(label, theme, theme.DefaultMonospaceFont(), theme.LabelStyle.FontColor)
	label.SetMargin(math.Spacing{L: 3, T: 3, R: 3, B: 3})
	return label
}

func (l *dirLabel) SetText(dir string) {
	if root, ok := fsroot(dir); ok {
		l.Label.SetText(root)
		return
	}
	if dir[len(dir)-1] != filepath.Separator {
		dir += string(filepath.Separator)
	}
	l.Label.SetText(dir)
}

func (l *dirLabel) Text() string {
	text := l.Label.Text()
	if root, ok := fsroot(text); ok {
		return root
	}
	if text == "" {
		log.Printf("This is odd.  We have an empty root that isn't considered a drive root.")
		return ""
	}
	if text[len(text)-1] == filepath.Separator {
		text = text[:len(text)-1]
	}
	return text
}
