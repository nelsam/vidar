// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package core

import (
	"log"
	"path/filepath"

	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/command/focus"
	"github.com/nelsam/vidar/core/internal/commander"
	"github.com/nelsam/vidar/core/internal/controller"
	"github.com/nelsam/vidar/core/internal/editor"
	"github.com/nelsam/vidar/core/internal/navigator"
	"github.com/nelsam/vidar/ui"
)

// Window represents the type of window that core needs in order to
// initialize.
type Window interface {
	Size() ui.Size
}

// Creator represents the pieces of the ui.Creator that core needs
// in order to initialize.
type Creator interface {
	Runner() ui.Runner
	LinearLayout(ui.Direction) (ui.Layout, error)
	Editor() (ui.TextContainer, error)
}

func Init(creator Creator, mainWindow Window, bindings []bind.Bindable) error {
	controller, err := controller.New(creator)
	if err != nil {
		return err
	}

	// Bindings should be added immediately after creating the commander,
	// since other types rely on the bindings having been bound.
	cmdr, err := commander.New(creator, controller)
	if err != nil {
		return err
	}
	cmdr.Push(bindings...)

	nav := navigator.New(creator)
	controller.SetNavigator(nav)

	editor := editor.New(creator, mainWindow, cmdr)
	controller.SetEditor(editor)

	projTree := navigator.NewProjectTree(cmdr, creator, mainWindow)
	projects := navigator.NewProjectsPane(cmdr, creator, projTree.Frame())

	nav.Add(projects)
	nav.Add(projTree)

	nav.Resize(mainWindow.Size().H)

	window.SetChild(cmdr)

	opener := cmdr.Bindable("focus-location").(*focus.Location)
	for _, file := range files {
		filepath, err := filepath.Abs(file)
		if err != nil {
			log.Printf("Failed to get path: %s", err)
		}
		cmdr.Execute(opener.For(focus.Path(filepath)))
	}
	return nil
}
