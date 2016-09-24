// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package settings

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
	"github.com/casimir/xdg-go"
	"github.com/spf13/viper"
)

const (
	LicenseHeaderFilename = ".license-header"

	projectsFilename = "projects"
	settingsFilename = "settings"
)

var (
	App              = xdg.App{Name: "vidar"}
	defaultConfigDir = filepath.Join(xdg.ConfigHome(), App.Name)
	projects         = viper.New()
	settings         = viper.New()
)

func init() {
	err := os.MkdirAll(defaultConfigDir, 0777)
	if err != nil {
		log.Printf("Error: Could not create config directory %s: %s", defaultConfigDir, err)
		return
	}
	projects.AddConfigPath(defaultConfigDir)
	projects.SetConfigName(projectsFilename)
	projects.SetTypeByDefaultValue(true)
	projects.SetDefault("projects", []Project(nil))

	settings.AddConfigPath(defaultConfigDir)
	settings.SetConfigName(settingsFilename)
	settings.SetTypeByDefaultValue(true)
	settings.SetDefault("fonts", []string(nil))
	err = settings.ReadInConfig()
	if _, unsupported := err.(viper.UnsupportedConfigError); unsupported {
		err = convertSettings()
	}
	if err != nil {
		log.Printf("Error reading settings: %s", err)
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

func Projects() (projs []Project) {
	err := projects.ReadInConfig()
	if _, unsupported := err.(viper.UnsupportedConfigError); unsupported {
		err = convertProjects()
	}
	if err != nil {
		log.Printf("Error parsing projects: %s", err)
	}
	if projs, ok := projects.Get("projects").([]Project); ok {
		return projs
	}
	// The following is to work around a bug in viper - it's not
	// actually inferring the []Project type.
	if err = projects.UnmarshalKey("projects", &projs); err != nil {
		log.Printf("Could not unmarshal projects: %s", err)
		return nil
	}
	return projs
}

func AddProject(project Project) {
	projects.Set("projects", append(projects.Get("projects").([]Project), project))
	if err := writeConfig(projects, projectsFilename); err != nil {
		log.Printf("Error updating projects file")
	}
}

func DesiredFonts() []string {
	return settings.Get("fonts").([]string)
}

func writeConfig(cfg *viper.Viper, fname string) error {
	f, err := os.Create(App.ConfigPath(fname + ".toml"))
	if err != nil {
		return fmt.Errorf("Could not create config file %s: %s", fname, err)
	}
	defer f.Close()
	encoder := toml.NewEncoder(f)
	if err := encoder.Encode(cfg.AllSettings()); err != nil {
		return fmt.Errorf("Could not marshal settings: %s", err)
	}
	return nil
}
