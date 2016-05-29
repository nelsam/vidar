// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package settings

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/nelsam/gxui"
	"github.com/spf13/viper"
)

const keysFilename = "keys"

var bindings = viper.New()

func init() {
	bindings.AddConfigPath(defaultConfigDir)
	bindings.SetConfigName(keysFilename)
	setDefaultBindings()
}

func Bindings(commandName string) (events []gxui.KeyboardEvent) {
	for event, action := range bindings.AllSettings() {
		if action == commandName {
			events = append(events, parseBinding(event)...)
		}
	}
	return events
}

func parseBinding(eventPattern string) []gxui.KeyboardEvent {
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

func setDefaultBindings() {
	bindings.SetDefault("Ctrl-Shift-N", "add-project")
	bindings.SetDefault("Ctrl-Shift-O", "open-project")
	bindings.SetDefault("Ctrl-O", "open-file")
	bindings.SetDefault("Ctrl-A", "select-all")
	bindings.SetDefault("Ctrl-S", "goimports, save-current-file")
	bindings.SetDefault("Ctrl-W", "close-current-tab")

	bindings.SetDefault("Ctrl-Z", "undo-last-edit")
	bindings.SetDefault("Ctrl-Shift-Z", "redo-next-edit")
	bindings.SetDefault("Ctrl-F", "find")
	bindings.SetDefault("Ctrl-C", "copy-selection")
	bindings.SetDefault("Ctrl-X", "cut-selection")
	bindings.SetDefault("Ctrl-V", "paste")
	bindings.SetDefault("Ctrl-Space", "show-suggestions")
	bindings.SetDefault("Ctrl-G", "goto-line")
	bindings.SetDefault("Ctrl-Shift-G", "goto-definition")
	bindings.SetDefault("Ctrl-Shift-L", "update-license")
	bindings.SetDefault("Ctrl-Shift-F", "goimports")
	bindings.SetDefault("Ctrl-/", "toggle-comments")

	bindings.SetDefault("Alt-H", "split-view-horizontally")
	bindings.SetDefault("Alt-V", "split-view-vertically")
	bindings.SetDefault("Ctrl-Tab", "next-tab")
	bindings.SetDefault("Ctrl-Shift-Tab", "prev-tab")

	bindings.SetDefault("Left", "prev-char")
	bindings.SetDefault("Ctrl-Left", "prev-word")
	bindings.SetDefault("Shift-Left", "select-prev-char")
	bindings.SetDefault("Ctrl-Shift-Left", "select-prev-word")
	bindings.SetDefault("Right", "next-char")
	bindings.SetDefault("Ctrl-Right", "next-word")
	bindings.SetDefault("Shift-Right", "select-next-char")
	bindings.SetDefault("Ctrl-Shift-Right", "select-next-word")
	bindings.SetDefault("Up", "prev-line")
	bindings.SetDefault("Shift-Up", "select-prev-line")
	bindings.SetDefault("Down", "next-line")
	bindings.SetDefault("Shift-Down", "select-next-line")
	bindings.SetDefault("End", "line-end")
	bindings.SetDefault("Shift-End", "select-to-line-end")
	bindings.SetDefault("Home", "line-start")
	bindings.SetDefault("Shift-Home", "select-to-line-start")
}

func writeBindings() error {
	f, err := os.Create(App.ConfigPath(keysFilename + ".toml"))
	if err != nil {
		return fmt.Errorf("Could not create config file: %s", err)
	}
	defer f.Close()
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(bindings.AllSettings()); err != nil {
		return fmt.Errorf("Could not marshal key bindings: %s", err)
	}
	return nil
}

func updateBindings() error {
	configNeedsUpdate := false
	for key, action := range bindings.AllSettings() {
		// The following were changed on May 30, 2016.  They should
		// be safe to remove by July, 2016.
		if action == "end-of-line" {
			bindings.Set(key, "line-end")
			configNeedsUpdate = true
		}
		if action == "beginning-of-line" {
			bindings.Set(key, "line-start")
			configNeedsUpdate = true
		}
	}
	if configNeedsUpdate {
		b, err := bindingConfigBytes()
		if err != nil {
			return err
		}
		b = bytes.Replace(b, []byte("end-of-line"), []byte("line-end"), -1)
		b = bytes.Replace(b, []byte("beginning-of-line"), []byte("line-start"), -1)
		return updateBindingConfigBytes(b)
	}
	return nil
}

func readBindings() error {
	err := bindings.ReadInConfig()
	if _, unsupported := err.(viper.UnsupportedConfigError); unsupported {
		err = writeBindings()
	}
	if err != nil {
		return err
	}
	return updateBindings()
}

func bindingConfigBytes() ([]byte, error) {
	f, err := os.Open(bindings.ConfigFileUsed())
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ioutil.ReadAll(f)
}

func updateBindingConfigBytes(bytes []byte) error {
	f, err := os.Create(bindings.ConfigFileUsed())
	if err != nil {
		return err
	}
	defer f.Close()
	written, err := f.Write(bytes)
	if err != nil {
		return err
	}
	if written < len(bytes) {
		return fmt.Errorf("Error: expected %d bytes to be written, got %d", len(bytes), written)
	}
	return nil
}
