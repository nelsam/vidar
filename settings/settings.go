// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package settings

import (
	"bytes"
	"io/ioutil"
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
	projectsFile = path.Join(settingsDir, "projects")
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
	projectsFile, err := os.Open(projectsFile)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		panic(err)
	}
	projectsBytes, err := ioutil.ReadAll(projectsFile)
	if err != nil {
		panic(err)
	}
	var projects []Project
	if err := yaml.Unmarshal(projectsBytes, &projects); err != nil {
		panic(err)
	}
	return projects
}

func AddProject(project Project) {
	projects, err := os.Create(projectsFile)
	if os.IsNotExist(err) {
		err = os.MkdirAll(path.Dir(projectsFile), 0777)
		if err != nil {
			panic(err)
		}
		projects, err = os.Create(projectsFile)
	}
	if err != nil {
		panic(err)
	}
	defer projects.Close()
	bytes, err := yaml.Marshal(append(Projects(), project))
	if err != nil {
		panic(err)
	}
	projects.Write(bytes)
}
