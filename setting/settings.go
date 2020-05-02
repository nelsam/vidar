// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package setting

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/OpenPeeDeeP/xdg"
	"github.com/nelsam/gxui"
	"github.com/nelsam/vidar/setting/config"
	"golang.org/x/image/font/gofont/gomono"
	"golang.org/x/image/font/gofont/gomonobold"
	"golang.org/x/image/font/gofont/gomonobolditalic"
	"golang.org/x/image/font/gofont/gomonoitalic"
)

const (
	// LicenseHeaderFilename is the file name to look for in projects.
	//
	// TODO: this belongs in the license header plugin, since not all
	// languages need it.
	LicenseHeaderFilename = ".license-header"

	// DefaultFontSize is the font size that will be used if no font
	// size settings are found in the config files.
	DefaultFontSize = 12

	projectsFilename = "projects"
	settingsFilename = "settings"
)

var (
	// App is an XDG application config.  It's exported so that plugins can load their own config
	// files from vidar's config directories.
	//
	// TODO: we should unexport this and provide functions to access its methods.  Allowing plugins
	// to assign to App is misleading and potentially dangerous.
	App              = xdg.New("", "vidar")
	defaultConfigDir = App.ConfigHome()
	projects         *config.Config
	settings         *config.Config

	// BuiltinFonts is a list of the fonts that we have built in to the
	// editor.  This is done so that vidar will always be able to start,
	// even if none of the fonts on a user's system are parseable.
	BuiltinFonts = map[string][]byte{
		"gomono":           gomono.TTF,
		"gomonobold":       gomonobold.TTF,
		"gomonoitalic":     gomonoitalic.TTF,
		"gomonobolditalic": gomonobolditalic.TTF,
	}
)

func init() {
	err := os.MkdirAll(defaultConfigDir, 0700)
	if err != nil {
		log.Printf("Error: Could not create config directory %s: %s", defaultConfigDir, err)
		return
	}
	projects, err = config.New(opener{}, projectsFilename, defaultConfigDir)
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		log.Printf("Error reading projects: %s", err)
	}
	projects.SetDefault("projects", []Project(nil))

	updateDeprecatedGopath(projects)

	settings, err = config.New(opener{}, settingsFilename, defaultConfigDir)
	if os.IsNotExist(err) {
		err = nil
	}
	if err != nil {
		log.Printf("Error reading settings: %s", err)
	}
	settings.SetDefault("fonts", []Font(nil))
}

func updateDeprecatedGopath(c *config.Config) error {
	write := false
	projList := projects.Get("projects").([]Project)
	for i, p := range projList {
		if p.Gopath == "" {
			continue
		}
		if p.Env == nil {
			p.Env = make(map[string]string)
		}
		p.Env["GOPATH"] = "=" + p.Gopath
		p.Env["PATH"] = filepath.Join(p.Gopath, "bin")
		p.Gopath = ""
		projList[i] = p
		write = true
	}
	if write {
		return c.Write()
	}
	return nil
}

type Font struct {
	Name string
	Size int
}

type Project struct {
	Name string
	Path string
	Env  map[string]string

	// Gopath is deprecated.  It is now merged into Env.
	// It's kept here for migration purposes.
	Gopath string `toml:",omitempty" json:",omitempty" yaml:",omitempty`
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
	for k, v := range p.Env {
		environ = addEnv(environ, k, v)
	}
	return environ
}

func addEnv(environ []string, key, value string) []string {
	if value == "" {
		return environ
	}
	envKey := key + "="
	replace := value[0] == '='
	if replace {
		value = value[1:]
	}
	value = parseEnv(value)
	for i, v := range environ {
		if !strings.HasPrefix(v, envKey) {
			continue
		}
		if replace {
			environ[i] = envKey + value
			return environ
		}
		environ[i] = fmt.Sprintf("%s%c%s", v, os.PathListSeparator, value)
		return environ
	}
	return append(environ, envKey+value)
}

// parseEnv does some very simple parsing to
// replace simple shell variable syntax with
// environment variables.
//
// TODO: allow use of environment variables
// set in the project's environment as well
// as the OS environment.
func parseEnv(orig string) string {
	v := orig
	if !strings.ContainsRune(v, '$') {
		return v
	}
	var b strings.Builder
	for ; len(v) > 0; v = v[1:] {
		if v[0] == '$' {
			name, newV, err := envVar(v)
			if err != nil {
				// TODO: handle these errors
				// gracefully
				return orig
			}
			v = newV
			b.WriteString(os.Getenv(name))
		}
		b.WriteByte(v[0])
	}
	return b.String()
}

// envVar reads the next environment variable,
// returning the variable name and the remaining
// text.  Only simple values will work - things
// like ${VAR:-someDefault} will not.
//
// Note: nested variables (i.e. using a variable's
// value as the name of another variable) are not
// supported.
func envVar(v string) (name, remaining string, err error) {
	if v[0] != '$' {
		return "", "", errors.New("string doesn't start with $")
	}
	v = v[1:]
	if v[0] == '{' {
		end := strings.IndexRune(v, '}')
		if end == -1 {
			return "", "", errors.New("could not find closing } in ${ syntax")
		}
		return v[1:end], v[end+1:], nil
	}
	end := strings.IndexFunc(v, func(r rune) bool {
		return !(unicode.IsDigit(r) || unicode.IsLetter(r))
	})
	if end == -1 {
		end = len(v)
	}
	return v[:end], v[end:], nil
}

func Projects() (projs []Project) {
	projs, ok := projects.Get("projects").([]Project)
	if !ok {
		return nil
	}
	return projs
}

func AddProject(project Project) {
	projects.Set("projects", append(Projects(), project))
	if err := projects.Write(); err != nil {
		log.Printf("Error updating projects file")
	}
}

func find(path, name string, extensions []string) (io.Reader, error) {
	d, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	i, err := d.Stat()
	if err != nil {
		return nil, err
	}
	if !i.IsDir() {
		return nil, fmt.Errorf("find called with a non-directory path")
	}
	infos, err := d.Readdir(-1)
	if err != nil {
		return nil, err
	}
	var lastErr error
	for _, i := range infos {
		if i.IsDir() {
			r, err := find(filepath.Join(path, i.Name()), name, extensions)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				lastErr = err
				continue
			}
			return r, nil
		}
		for _, e := range extensions {
			if i.Name() != fmt.Sprintf("%s.%s", name, e) {
				continue
			}
			r, err := os.Open(filepath.Join(path, i.Name()))
			if err != nil {
				lastErr = err
				continue
			}
			return r, nil
		}
	}
	if lastErr != nil {
		return nil, err
	}
	return nil, os.ErrNotExist
}

func loadFont(font string) (io.Reader, error) {
	if b, ok := BuiltinFonts[font]; ok {
		return bytes.NewBuffer(b), nil
	}
	ext := []string{"ttf", "otf"}
	var lastErr error
	for _, p := range fontPaths {
		r, err := find(p, font, ext)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			lastErr = err
			continue
		}
		return r, nil
	}
	if lastErr != nil {
		return nil, lastErr
	}
	return nil, os.ErrNotExist
}

// PrefFont returns the most preferred font found on the system.
func PrefFont(d gxui.Driver) gxui.Font {
	fonts, ok := settings.Get("fonts").([]Font)
	if !ok {
		return parseDefaultFont(d)
	}
	for _, font := range fonts {
		r, err := loadFont(font.Name)
		if err != nil {
			log.Printf("Failed to load font %s: %s", font.Name, err)
			continue
		}
		f, err := parseFont(d, r, font.Size)
		if err != nil {
			log.Printf("Failed to parse font %s: %s", font.Name, err)
			continue
		}
		return f
	}
	return parseDefaultFont(d)
}

func parseDefaultFont(d gxui.Driver) gxui.Font {
	f, err := parseFont(d, bytes.NewBuffer(gomono.TTF), DefaultFontSize)
	if err != nil {
		// This is a well-tested font that should never fail to parse.
		panic(fmt.Errorf("failed to parse default font: %s", err))
	}
	return f
}

func parseFont(d gxui.Driver, f io.Reader, size int) (gxui.Font, error) {
	if c, ok := f.(io.Closer); ok {
		defer c.Close()
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	font, err := d.CreateFont(b, size)
	if err != nil {
		return nil, err
	}
	return font, nil
}
