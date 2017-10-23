// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package settings

import (
	"log"
	"os"
	"path/filepath"
)

const pluginsDirname = "plugins"

func Plugins() []string {
	pluginsPath := App.DataPath(pluginsDirname)
	dir, err := os.Open(pluginsPath)
	if err == os.ErrNotExist {
		return nil
	}
	if err != nil {
		log.Printf("Failed to read directory %s: %s", pluginsPath, err)
		return nil
	}

	finfos, err := dir.Readdir(-1)
	if err != nil {
		log.Printf("Failed to list directory contents at %s: %s", pluginsPath, err)
		return nil
	}

	paths := make([]string, 0, len(finfos))
	for _, finfo := range finfos {
		paths = append(paths, filepath.Join(pluginsPath, finfo.Name()))
	}
	return paths
}
