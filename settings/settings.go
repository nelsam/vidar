// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.
package settings

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
)

var (
	settingsDir  = path.Join(os.Getenv("HOME"), ".config", "vidar")
	projectsFile = path.Join(settingsDir, "projects")

	// AssetsDir is the directory where assets (images and the like)
	// are stored.  It should be set either at compile time or via
	// command line flags.
	AssetsDir string
)

type Project struct {
	Name, Path string
}

func (p Project) Line() []byte {
	return []byte(fmt.Sprintf("%s: %s", p.Name, p.Path))
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
	for _, line := range bytes.Split(projectsBytes, []byte{'\n'}) {
		if len(line) > 0 {
			projects = append(projects, fromLine(line))
		}
	}
	return projects
}

func AddProject(project Project) {
	projects, err := os.OpenFile(projectsFile, os.O_APPEND, os.ModeAppend)
	if os.IsNotExist(err) {
		projects, err = os.Create(projectsFile)
		if os.IsNotExist(err) {
			err = os.MkdirAll(path.Dir(projectsFile), 0777)
			if err != nil {
				panic(err)
			}
			projects, err = os.Create(projectsFile)
		}
	}
	if err != nil {
		panic(err)
	}
	defer projects.Close()
	projects.Write(project.Line())
}
