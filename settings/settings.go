// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package settings

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/casimir/xdg-go"
	"gopkg.in/yaml.v2"
)

const (
	LicenseHeaderFilename = ".license-header"
)

var (
	App              = xdg.App{Name: "vidar"}
	defaultConfigDir = filepath.Join(xdg.ConfigHome(), App.Name)
	projectsPath     = App.ConfigPath("projects")
	settingsPath     = App.ConfigPath("settings")

	appSettings settings
)

func init() {
	err := os.MkdirAll(filepath.Dir(projectsPath), 0777)
	if err != nil {
		log.Printf("Error: Could not create config directory %s: %s", filepath.Dir(projectsPath), err)
		return
	}
	f, err := os.Open(settingsPath)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		log.Printf("Could not open settings file: %s", err)
		return
	}
	defer f.Close()
	settingsBytes, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Could not read settings file: %s", err)
		return
	}
	if err := yaml.Unmarshal(settingsBytes, &appSettings); err != nil {
		log.Printf("Error: Could not parse %s as yaml: %s", settingsPath, err)
		return
	}
}

type Project struct {
	Name   string
	Path   string
	Gopath string
}

func (p Project) LicenseHeader() string {
	f, err := os.Open(filepath.Join(p.Path, LicenseHeaderFilename))
	if os.IsNotExist(err) {
		return ""
	}
	if err != nil {
		log.Printf("Error opening license header file: %s", err)
		return ""
	}
	defer f.Close()
	header, err := ioutil.ReadAll(f)
	if err != nil {
		log.Printf("Error reading license header file: %s", err)
		return ""
	}
	return string(header)
}

func (p Project) String() string {
	return p.Name
}

func fromLine(line []byte) Project {
	parts := bytes.SplitN(line, []byte{':'}, 2)
	return Project{
		Name: string(bytes.TrimSpace(parts[0])),
		Path: string(bytes.TrimSpace(parts[1])),
	}
}

func Projects() []Project {
	projectsFile, err := os.Open(projectsPath)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		log.Printf("Error: Could not open %s: %s", projectsPath, err)
		return nil
	}
	projectsBytes, err := ioutil.ReadAll(projectsFile)
	if err != nil {
		log.Printf("Error: Could not read %s: %s", projectsPath, err)
		return nil
	}
	var projects []Project
	if err := yaml.Unmarshal(projectsBytes, &projects); err != nil {
		log.Printf("Error: Could not parse %s as yaml: %s", projectsPath, err)
		return nil
	}
	return projects
}

func AddProject(project Project) {
	bytes, err := yaml.Marshal(append(Projects(), project))
	if err != nil {
		log.Printf("Error: Could not convert project to yaml: %s", err)
		return
	}
	projects, err := os.Create(projectsPath)
	if err != nil {
		log.Printf("Could not open %s for writing: %s", projectsPath, err)
		return
	}
	defer projects.Close()
	projects.Write(bytes)
}

func DesiredFonts() []string {
	return appSettings.Fonts
}

type settings struct {
	Fonts []string
}
