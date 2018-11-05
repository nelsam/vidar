// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"bytes"
	"image"
	"image/draw"
	"log"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/asset"
	"github.com/nfnt/resize"

	// Supported image types
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
)

func createIconButton(driver gxui.Driver, theme gxui.Theme, iconPath string) gxui.Button {
	button := theme.CreateButton()
	button.SetType(gxui.PushButton)

	fileBytes, err := asset.Asset(iconPath)
	if err != nil {
		log.Printf("Error: Failed to read asset %s: %s", iconPath, err)
		return button
	}
	f := bytes.NewBuffer(fileBytes)
	src, _, err := image.Decode(f)
	if err != nil {
		log.Printf("Error: Failed to decode image %s: %s", iconPath, err)
		return button
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
