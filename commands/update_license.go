// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"strings"

	"github.com/nelsam/gxui"
)

type LicenseHeaderUpdate struct {
	statusKeeper
}

func NewLicenseHeaderUpdate(theme gxui.Theme) *LicenseHeaderUpdate {
	return &LicenseHeaderUpdate{
		statusKeeper: statusKeeper{theme: theme},
	}
}

func (u *LicenseHeaderUpdate) Name() string {
	return "update-license"
}

func (u *LicenseHeaderUpdate) Menu() string {
	return "Edit"
}

func (u *LicenseHeaderUpdate) Exec(on interface{}) (executed, consume bool) {
	finder, ok := on.(ProjectFinder)
	if !ok {
		return false, false
	}
	edit := u.LicenseEdit(finder)
	if edit == nil {
		return true, true
	}
	editor := finder.CurrentEditor()
	oldRunes := editor.Controller().TextRunes()
	runes := make([]rune, 0, len(oldRunes)+edit.Delta)
	runes = append(runes, edit.New...)
	runes = append(runes, oldRunes[len(edit.Old):]...)

	editor.Controller().SetTextEdits(runes, []gxui.TextBoxEdit{*edit})
	return true, true
}

func (u *LicenseHeaderUpdate) LicenseEdit(finder ProjectFinder) *gxui.TextBoxEdit {
	editor := finder.CurrentEditor()
	if editor == nil {
		return nil
	}
	licenseHeader := strings.TrimSpace(finder.Project().LicenseHeader())
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
		u.info = "license is already set correctly"
		return nil
	}
	return &gxui.TextBoxEdit{
		At:    0,
		Delta: len(licenseHeader) - len(commentText),
		Old:   []rune(commentText),
		New:   []rune(licenseHeader),
	}
}
