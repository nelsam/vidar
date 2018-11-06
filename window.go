// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package main

import (
	"bytes"
	"fmt"
	"image"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/asset"
)

// icon returns the image.Image to be used as vidar's icon.
func icon() image.Image {
	r := bytes.NewReader(asset.MustAsset("icon.png"))
	i, _, err := image.Decode(r)
	if err != nil {
		panic(fmt.Errorf("could not decode png: %s", err))
	}
	return i
}

// window is a slightly modified gxui.Window that we use to wrap up
// some child types.
type window struct {
	gxui.Window
	child interface{}
}

func newWindow(t gxui.Theme) *window {
	w := &window{
		Window: t.CreateWindow(1600, 800, "Vidar Text Editor"),
	}
	w.SetIcon(icon())
	return w
}

func (w *window) Elements() []interface{} {
	return []interface{}{w.child}
}
