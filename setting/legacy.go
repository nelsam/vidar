// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package setting

import (
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"
)

func convertProjects() error {
	projectsPath := filepath.Join(defaultConfigDir, projectsFilename)
	projectsFile, err := os.Open(projectsPath)
	if os.IsNotExist(err) {
		return nil
	}
	// If the legacy file exists, we want to remove it regardless of any
	// problems past this point.
	defer os.Remove(projectsPath)
	if err != nil {
		return err
	}
	projectsBytes, err := ioutil.ReadAll(projectsFile)
	if err != nil {
		return err
	}
	var projs []Project
	if err := yaml.Unmarshal(projectsBytes, &projs); err != nil {
		return err
	}
	projects.Set("projects", projs)
	return writeConfig(projects, projectsFilename)
}

func convertSettings() error {
	settingsPath := filepath.Join(defaultConfigDir, settingsFilename)
	f, err := os.Open(settingsPath)
	if os.IsNotExist(err) {
		return nil
	}
	// As with projects, remove the legacy file no matter what.
	defer os.Remove(settingsPath)
	if err != nil {
		return err
	}
	defer f.Close()
	settingsBytes, err := ioutil.ReadAll(f)
	if err != nil {
		return err
	}
	var appSettings map[string][]string
	if err := yaml.Unmarshal(settingsBytes, &appSettings); err != nil {
		return err
	}
	settings.Set("fonts", appSettings["fonts"])
	return writeConfig(settings, settingsFilename)
}
