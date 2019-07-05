// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package setting

import (
	"io"
	"log"
	"os"
	"strings"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/setting/config"
)

const keysFilename = "keys"

var bindings *config.Config

func init() {
	var err error
	bindings, err = config.New(opener{}, keysFilename, defaultConfigDir)
	if err != nil {
		log.Printf("Error reading key bindings: %s", err)
	}
}

type opener struct{}

func (o opener) Open(path string) (io.ReadCloser, error) {
	return os.Open(path)
}

func (o opener) Create(path string) (io.WriteCloser, error) {
	return os.Create(path)
}

func Bindings(commandName string) (events []gxui.KeyboardEvent) {
	for _, event := range bindings.Keys() {
		if bindings.Get(event) == commandName {
			events = append(events, parseBinding(event)...)
		}
	}
	return events
}

func parseBinding(eventPattern string) []gxui.KeyboardEvent {
	// TODO: Move this logic to input.Handler so that other handlers can define
	// their own keybinding format.
	eventPattern = strings.ToLower(eventPattern)
	keys := strings.Split(eventPattern, "-")
	modifiers, key := keys[:len(keys)-1], keys[len(keys)-1]
	var event gxui.KeyboardEvent
	for _, key := range modifiers {
		switch key {
		case "ctrl", "cmd":
			event.Modifier |= gxui.ModControl
		case "alt":
			event.Modifier |= gxui.ModAlt
		case "shift":
			event.Modifier |= gxui.ModShift
		case "super":
			log.Printf("Error: %s: Super cannot be bound directly; use ctrl or cmd instead.", eventPattern)
			return nil
		default:
			log.Printf("Error parsing key bindings: Modifier %s not understood", key)
		}
	}
	for k := gxui.KeyboardKey(0); k < gxui.KeyLast; k++ {
		if strings.ToLower(k.String()) == key {
			event.Key = k
			events := []gxui.KeyboardEvent{event}
			if event.Modifier.Control() {
				// Make ctrl and cmd mirror each other, for those of us who
				// need to switch between OS X and linux on a regular basis.
				event.Modifier &^= gxui.ModControl
				event.Modifier |= gxui.ModSuper
				events = append(events, event)
			}
			return events
		}
	}
	log.Printf("Error parsing key bindings: Key %s not understood", key)
	return nil
}

func SetDefaultBindings(cmds ...bind.Command) {
	for _, c := range cmds {
		defaults := c.Defaults()
		for _, d := range defaults {
			bindings.SetDefault(d.String(), c.Name())
		}
	}
	bindings.Write()
}
