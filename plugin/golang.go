// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

// +build !linux !go1.8

package plugin

import (
	"strings"

	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/plugin/comments"
	"github.com/nelsam/vidar/plugin/godef"
	"github.com/nelsam/vidar/plugin/goimports"
	"github.com/nelsam/vidar/plugin/gosyntax"
	"github.com/nelsam/vidar/plugin/license"
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

func (h GolangHook) FileBindables(path string) []bind.Bindable {
	if !strings.HasSuffix(path, ".go") {
		return nil
	}
	return []bind.Bindable{
		comments.NewToggle(),
		godef.New(h.Theme),
		goimports.New(h.Theme),
		goimports.OnSave{},
		gosyntax.New(),
		license.NewHeaderUpdate(h.Theme),
	}
}
