package navigator

import (
	"bytes"
	"image"
	"image/draw"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/assets"
	"github.com/nfnt/resize"

	// Supported image types
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func createIconButton(driver gxui.Driver, theme gxui.Theme, iconPath string) gxui.Button {
	button := theme.CreateButton()
	button.SetType(gxui.PushButton)

	fileBytes, err := assets.Asset(iconPath)
	if err != nil {
		panic(err)
	}
	f := bytes.NewBuffer(fileBytes)
	src, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	src = resize.Resize(24, 24, src, resize.Bilinear)

	rgba := image.NewRGBA(src.Bounds())
	draw.Draw(rgba, src.Bounds(), src, image.ZP, draw.Src)
	texture := driver.CreateTexture(rgba, 1)

	icon := theme.CreateImage()
	icon.SetTexture(texture)
	button.AddChild(icon)
	return button
}
