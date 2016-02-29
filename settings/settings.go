// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package settings

import (
	"bytes"
	"io/ioutil"
	"log"
	"os"
	"path"

	"gopkg.in/yaml.v2"
)

var (
	DefaultProject = Project{
		Name:   "*default*",
		Path:   "/",
		Gopath: os.Getenv("GOPATH"),
	}
	settingsDir  = path.Join(os.Getenv("HOME"), ".config", "vidar")
	projectsPath = path.Join(settingsDir, "projects")
)

type Project struct {
	Name   string
	Path   string
	Gopath string
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
	if os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(projectsPath), 0777)
		if err != nil {
			log.Printf("Error: Could not create %s: %s", path.Dir(projectsPath), err)
			return
		}
		projects, err = os.Create(projectsPath)
	}
	if err != nil {
		log.Printf("Could not open %s for writing: %s", projectsPath, err)
		return
	}
	defer projects.Close()
	projects.Write(bytes)
}
