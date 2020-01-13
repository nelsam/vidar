// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package license

import (
	"fmt"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/plugin/status"
	"github.com/nelsam/vidar/setting"
)

type Projecter interface {
	Project() setting.Project
}

type Applier interface {
	Apply(input.Editor, ...input.Edit)
}

type Controller interface {
	LineCount() int
	LineStart(int) int
	Line(int) string
}

type HeaderUpdate struct {
	status.General

	editor    input.Editor
	applier   Applier
	projecter Projecter
	ctrl      Controller
}

func NewHeaderUpdate(theme gxui.Theme) *HeaderUpdate {
	u := &HeaderUpdate{}
	u.Theme = theme
	return u
}

func (u *HeaderUpdate) Name() string {
	return "update-license"
}

func (u *HeaderUpdate) Menu() string {
	return "License"
}

func (u *HeaderUpdate) Defaults() []fmt.Stringer {
	return []fmt.Stringer{gxui.KeyboardEvent{
		Modifier: gxui.ModControl | gxui.ModShift,
		Key:      gxui.KeyL,
	}}
}

func (u *HeaderUpdate) Store(target interface{}) bind.Status {
	switch src := target.(type) {
	case Projecter:
		u.projecter = src
	case Applier:
		u.applier = src
	case input.Editor:
		u.editor = src
	case Controller:
		u.ctrl = src
	}
	if u.projecter != nil && u.applier != nil && u.editor != nil && u.ctrl != nil {
		return bind.Done
	}
	return bind.Waiting
}

func (u *HeaderUpdate) Reset() {
	u.projecter = nil
	u.applier = nil
	u.editor = nil
	u.ctrl = nil
}

func (u *HeaderUpdate) Exec() error {
	edit := u.LicenseEdit()
	if edit == nil {
		return nil
	}
	u.applier.Apply(u.editor, *edit)
	return nil
}

func (u *HeaderUpdate) LicenseEdit() *input.Edit {
	licenseHeader := strings.TrimSpace(u.projecter.Project().LicenseHeader())
	if licenseHeader != "" {
		licenseHeader += "\n\n"
	}

	// License headers should be the first comment block on a file,
	// should have an empty line following them, and should not
	// contain build tag comments
	firstCommentBlockEnd := 0
	numberOfLines := u.ctrl.LineCount()
	for i := 0; i < numberOfLines; i++ {
		line := u.ctrl.Line(i)
		if !strings.HasPrefix(string(line), "//") && i+1 != numberOfLines {
			if len(line) == 0 {
				firstCommentBlockEnd = u.ctrl.LineStart(i + 1)
			}
			break
		}
	}
	commentText := u.editor.Text()[:firstCommentBlockEnd]
	if strings.Contains(commentText, "// +build ") {
		commentText = ""
	}
	if commentText == licenseHeader {
		u.Info = "license is already set correctly"
		return nil
	}
	return &input.Edit{
		At:  0,
		Old: []rune(commentText),
		New: []rune(licenseHeader),
	}
}
