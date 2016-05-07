// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package commands

import (
	"log"
	"strings"

	"github.com/nelsam/gxui"
)

type LicenseHeaderUpdate struct {
}

func NewLicenseHeaderUpdate() *LicenseHeaderUpdate {
	return &LicenseHeaderUpdate{}
}

func (u *LicenseHeaderUpdate) Start(gxui.Control) gxui.Control {
	return nil
}

func (u *LicenseHeaderUpdate) Name() string {
	return "goimports"
}

func (u *LicenseHeaderUpdate) Next() gxui.Focusable {
	return nil
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
	runes := editor.Controller().TextRunes()
	runes = append(runes[:len(edit.New)], runes[len(edit.Old):]...)
	copy(runes, edit.New)

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
	for i := 0; i < editor.Controller().LineCount(); i++ {
		line := editor.Controller().LineRunes(i)
		if !strings.HasPrefix(string(line), "//") {
			if len(line) == 0 {
				firstCommentBlockEnd = editor.Controller().LineStart(i + 1)
			}
			break
		}
	}
	commentText := editor.Text()[:firstCommentBlockEnd]
	if strings.Contains(commentText, "// +build ") {
		commentText = ""
		firstCommentBlockEnd = 0
	}
	if commentText == licenseHeader {
		log.Printf("License header update: license is already set correctly")
		return nil
	}
	return &gxui.TextBoxEdit{
		At:    0,
		Delta: len(licenseHeader) - len(commentText),
		Old:   []rune(commentText),
		New:   []rune(licenseHeader),
	}
}
