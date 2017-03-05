// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package plugins

import (
	"strings"

	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander"
	"github.com/nelsam/vidar/plugins/comments"
	"github.com/nelsam/vidar/plugins/godef"
	"github.com/nelsam/vidar/plugins/goimports"
	"github.com/nelsam/vidar/plugins/license"
)

type GolangHook struct {
	Theme *basic.Theme
}

func (h GolangHook) Name() string {
	return "golang-hook"
}

func (h GolangHook) CommandName() string {
	return "open-file"
}

func (h GolangHook) FileBindables(path string) []commander.Bindable {
	if !strings.HasSuffix(path, ".go") {
		return nil
	}
	return []commander.Bindable{
		comments.NewToggle(),
		godef.New(h.Theme),
		goimports.New(h.Theme),
		goimports.OnSave{},
		license.NewHeaderUpdate(h.Theme),
	}
}
