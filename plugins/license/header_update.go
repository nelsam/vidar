// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package license

import (
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/settings"
)

type OpenProject interface {
	CurrentEditor() *editor.CodeEditor
	Project() settings.Project
}

type HeaderUpdate struct {
	commander.GenericStatuser
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
	return "Edit"
}

func (u *HeaderUpdate) Exec(on interface{}) (executed, consume bool) {
	proj, ok := on.(OpenProject)
	if !ok {
		return false, false
	}
	edit := u.LicenseEdit(proj)
	if edit == nil {
		return true, true
	}
	editor := proj.CurrentEditor()
	oldRunes := editor.Controller().TextRunes()
	runes := make([]rune, 0, len(oldRunes)+edit.Delta)
	runes = append(runes, edit.New...)
	runes = append(runes, oldRunes[len(edit.Old):]...)

	editor.Controller().SetTextEdits(runes, []gxui.TextBoxEdit{*edit})
	return true, true
}

func (u *HeaderUpdate) LicenseEdit(proj OpenProject) *gxui.TextBoxEdit {
	editor := proj.CurrentEditor()
	if editor == nil {
		return nil
	}
	licenseHeader := strings.TrimSpace(proj.Project().LicenseHeader())
	if licenseHeader != "" {
		licenseHeader += "\n\n"
	}

	// License headers should be the first comment block on a file,
	// should have an empty line following them, and should not
	// contain build tag comments
	firstCommentBlockEnd := 0
	numberOfLines := editor.Controller().LineCount()
	for i := 0; i < numberOfLines; i++ {
		line := editor.Controller().LineRunes(i)
		if !strings.HasPrefix(string(line), "//") && i+1 != numberOfLines {
			if len(line) == 0 {
				firstCommentBlockEnd = editor.Controller().LineStart(i + 1)
			}
			break
		}
	}
	commentText := editor.Text()[:firstCommentBlockEnd]
	if strings.Contains(commentText, "// +build ") {
		commentText = ""
	}
	if commentText == licenseHeader {
		u.Info = "license is already set correctly"
		return nil
	}
	return &gxui.TextBoxEdit{
		At:    0,
		Delta: len(licenseHeader) - len(commentText),
		Old:   []rune(commentText),
		New:   []rune(licenseHeader),
	}
}
