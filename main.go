// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"log"

	"github.com/nelsam/vidar/bind"
	"github.com/nelsam/vidar/command"
	"github.com/nelsam/vidar/core/input"

	"github.com/nelsam/vidar/core"
	"github.com/nelsam/vidar/plugin"
	"github.com/nelsam/vidar/ui"
	"github.com/spf13/cobra"
)

var (
	cmd   *cobra.Command
	files []string
)

func init() {
	cmd = &cobra.Command{
		Use:   "vidar [files...]",
		Short: "An experimental Go editor",
		Long: "An editor for Go code, still in its infancy.  " +
			"Basic editing of Go code is mostly complete, but " +
			"panics still happen and can result in the loss of " +
			"unsaved work.",
		Run: start,
	}
	// TODO: allow overrides for some settings via flags, e.g.
	// --ui=gtk or --ui=shiny.
}

func main() {
	cmd.Execute()
}

func start(cmd *cobra.Command, files []string) {
	bindings := []bind.Bindable{input.New(driver, cmdr)}
	bindings = append(bindings, core.Bindables(cmdr, driver, gTheme)...)
	bindings = append(bindings, command.Bindables(cmdr, driver, gTheme)...)
	bindings = append(bindings, plugin.Bindables(cmdr, driver, gTheme)...)

	var creator ui.Creator
	for _, b := range bindings {
		c, ok := b.(ui.Creator)
		if !ok {
			continue
		}
		creator = c
		// We don't break here because we want to prioritize UIs that were loaded
		// later (e.g. plugins).
		//
		// TODO: decide on a way for the user to enable/disable specific plugins.
	}
	if err := creator.Start(); err != nil {
		log.Fatalf("Could not start UI %#v: %s", creator, err)
		// TODO: should we fall back to other UIs if they exist?
	}
	defer creator.Quit()

	// TODO: allow multiple windows.
	core.Start(creator, creator.Window(), bindings)
	creator.Wait()
}
