// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package settings

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/casimir/xdg-go"
	"github.com/nelsam/gxui"
	"github.com/spf13/viper"
)

const keysFilename = "keys"

var Keybindings = viper.New()

func init() {
	Keybindings.AddConfigPath(filepath.Join(xdg.ConfigHome(), App.Name))
	Keybindings.SetConfigName(keysFilename)
	setDefaultBindings()
	err := Keybindings.ReadInConfig()
	if _, unsupported := err.(viper.UnsupportedConfigError); unsupported {
		writeDefaultBindings()
		err = Keybindings.ReadInConfig()
	}
	if err != nil {
		panic(fmt.Errorf("Fatal error reading keybindings: %s", err))
	}
}

func Bindings(commandName string) (bindings []gxui.KeyboardEvent) {
	for event, action := range Keybindings.AllSettings() {
		if action == commandName {
			bindings = append(bindings, parseBinding(event)...)
		}
	}
	return bindings
}

func parseBinding(eventPattern string) []gxui.KeyboardEvent {
	eventPattern = strings.ToUpper(eventPattern)
	keys := strings.Split(eventPattern, "-")
	modifiers, key := keys[:len(keys)-1], keys[len(keys)-1]
	var event gxui.KeyboardEvent
	for _, key := range modifiers {
		switch key {
		case "CTRL", "CMD":
			event.Modifier |= gxui.ModControl
		case "ALT":
			event.Modifier |= gxui.ModAlt
		case "SHIFT":
			event.Modifier |= gxui.ModShift
		case "SUPER":
			log.Printf("Error: %s: Super cannot be bound directly; use ctrl or cmd instead.", eventPattern)
			return nil
		default:
			log.Printf("Error parsing key bindings: Modifier %s not understood", key)
		}
	}
	// TODO: This is making an assumption about keys supported in gxui
	// and the order they are defined in.  I'd rather not do that.
	for k := gxui.KeyboardKey(0); k < gxui.KeyLast; k++ {
		if k.String() == key {
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

func setDefaultBindings() {
	Keybindings.SetDefault("Ctrl-Shift-N", "add-project")
	Keybindings.SetDefault("Ctrl-Shift-O", "open-project")
	Keybindings.SetDefault("Ctrl-O", "open-file")
	Keybindings.SetDefault("Ctrl-S", "goimports, save-current-file")
	Keybindings.SetDefault("Ctrl-W", "close-current-tab")

	Keybindings.SetDefault("Ctrl-Z", "undo-last-edit")
	Keybindings.SetDefault("Ctrl-Shift-Z", "redo-next-edit")
	Keybindings.SetDefault("Ctrl-F", "find")
	Keybindings.SetDefault("Ctrl-C", "copy-selection")
	Keybindings.SetDefault("Ctrl-X", "cut-selection")
	Keybindings.SetDefault("Ctrl-V", "paste")
	Keybindings.SetDefault("Ctrl-G", "goto-line")
	Keybindings.SetDefault("Ctrl-Shift-G", "goto-definition")
	Keybindings.SetDefault("Ctrl-Shift-L", "update-license")
	Keybindings.SetDefault("Ctrl-Shift-F", "goimports")
	Keybindings.SetDefault("Ctrl-/", "toggle-comments")

	Keybindings.SetDefault("Alt-H", "split-view-horizontally")
	Keybindings.SetDefault("Alt-V", "split-view-vertically")
}

func writeDefaultBindings() {
	f, err := os.Create(App.ConfigPath(keysFilename + ".toml"))
	if err != nil {
		panic(fmt.Errorf("Could not create config file: %s", err))
	}
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(Keybindings.AllSettings()); err != nil {
		panic(fmt.Errorf("Could not marshal default key bindings: %s", err))
	}
}
