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
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/casimir/xdg-go"
	"github.com/spf13/viper"
)

const (
	LicenseHeaderFilename = ".license-header"
	DefaultFontSize       = 12

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
	settings.SetDefault("fonts", []Font(nil))
	readSettings()
}

func readSettings() {
	err := settings.ReadInConfig()
	if _, notFound := err.(viper.ConfigFileNotFoundError); notFound {
		return
	}
	if _, unsupported := err.(viper.UnsupportedConfigError); unsupported {
		err = convertSettings()
	}
	if err != nil {
		log.Printf("Error reading settings: %s", err)
	}
}

type Font struct {
	Name string
	Size int
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

func (p Project) Environ() []string {
	environ := os.Environ()
	if p.Gopath == "" {
		return environ
	}
	environ = addEnv(environ, "GOPATH", p.Gopath, true)
	return addEnv(environ, "PATH", filepath.Join(p.Gopath, "bin"), false)
}

func addEnv(environ []string, key, value string, replace bool) []string {
	envKey := key + "="
	env := envKey + value
	for i, v := range environ {
		if !strings.HasPrefix(v, envKey) {
			continue
		}
		if !replace {
			env = fmt.Sprintf("%s%c%s", v, os.PathSeparator, value)
		}
		environ[i] = env
		return environ
	}
	return append(environ, env)
}

func Projects() (projs []Project) {
	err := projects.ReadInConfig()
	if _, notFound := err.(viper.ConfigFileNotFoundError); notFound {
		return nil
	}
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
	projects.Set("projects", append(Projects(), project))
	if err := writeConfig(projects, projectsFilename); err != nil {
		log.Printf("Error updating projects file")
	}
}

func DesiredFonts() (fonts []Font) {
	if fonts, ok := settings.Get("fonts").([]Font); ok {
		return fonts
	}
	// The following is to work around the same bug in viper as
	// mentioned in Projects()
	if err := settings.UnmarshalKey("fonts", &fonts); err != nil {
		convertFonts()
		return settings.Get("fonts").([]Font)
	}
	return fonts
}

func convertFonts() {
	oldFonts, ok := settings.Get("fonts").([]interface{})
	if !ok {
		settings.Set("fonts", []Font(nil))
		return
	}
	newFonts := make([]Font, 0, len(oldFonts))
	for _, f := range oldFonts {
		newFonts = append(newFonts, Font{
			Name: f.(string),
			Size: DefaultFontSize,
		})
	}
	settings.Set("fonts", newFonts)
	if err := writeConfig(settings, settingsFilename); err != nil {
		log.Printf("Failed to update fonts: %s", err)
	}
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
