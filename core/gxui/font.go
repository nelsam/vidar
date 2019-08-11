// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package gxui

import (
	"io"
	"io/ioutil"
	"log"

	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/setting"
	"github.com/tmc/fonts"
)

func font(driver gxui.Driver) gxui.Font {
	desiredFonts := setting.DesiredFonts()
	if len(desiredFonts) == 0 {
		return nil
	}
	var (
		font       setting.Font
		fontReader io.Reader
		err        error
	)
	for _, font = range desiredFonts {
		fontReader, err = fonts.Load(font.Name)
		if err == nil {
			break
		}
	}
	if err != nil {
		return nil
	}
	if closer, ok := fontReader.(io.Closer); ok {
		defer closer.Close()
	}
	fontBytes, err := ioutil.ReadAll(fontReader)
	if err != nil {
		log.Printf("Failed to read font file: %s", err)
		return nil
	}
	gFont, err := driver.CreateFont(fontBytes, font.Size)
	if err != nil {
		log.Printf("Could not parse font: %s", err)
		return nil
	}
	return gFont
}
