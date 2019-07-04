// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package setting

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type unmarshaller func(io.Reader, interface{}) error

func decJSON(r io.Reader, t interface{}) error {
	return json.NewDecoder(r).Decode(t)
}

func decYAML(r io.Reader, t interface{}) error {
	return yaml.NewDecoder(r).Decode(t)
}

func decTOML(r io.Reader, t interface{}) error {
	_, err := toml.DecodeReader(r, t)
	return err
}

// config is a type to mirror viper.Config, but with better custom
// type support.  This stems from the closure of
// github.com/spf13/viper#271.
type config struct {
	name    string
	dirs    []string
	formats map[string]unmarshaller

	// TODO: use an fsw.Watcher to watch for filesystem changes and
	// reload.
	data map[string]interface{}
}

func newConfig(name string, dirs ...string) (*config, error) {
	c := &config{
		name: name,
		dirs: dirs,
		formats: map[string]unmarshaller{
			"toml": decTOML,
			"yaml": decYAML,
			"yml":  decYAML,
			"json": decJSON,
		},
	}
	return c, c.load()
}

func (c *config) bestFile() (io.ReadCloser, unmarshaller, error) {
	for _, d := range c.dirs {
		for ext, unm := range c.formats {
			f, err := os.Open(filepath.Join(d, fmt.Sprintf("%s.%s", c.name, ext)))
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, nil, err
			}
			return f, unm, nil
		}
	}
	return nil, nil, os.ErrNotExist
}

func (c *config) load() error {
	f, unm, err := c.bestFile()
	if err != nil {
		return err
	}
	defer f.Close()
	if err := unm(f, &c.data); err != nil {
		return err
	}
	for k, v := range c.data {
		lk := strings.ToLower(k)
		if lk == k {
			continue
		}
		if _, ok := c.data[lk]; ok {
			return fmt.Errorf("duplicate keys %s and %s found; configs are case insensitive", k, lk)
		}
		c.data[lk] = v
		delete(c.data, k)
	}
	return nil
}

func (c *config) setDefault(k string, v interface{}) {
	k = strings.ToLower(k)
	if d, ok := c.data[k]; ok {
		c.data[k] = convert(reflect.ValueOf(d), reflect.TypeOf(v)).Interface()
		return
	}
	c.set(k, v)
}

func (c *config) set(k string, v interface{}) {
	k = strings.ToLower(k)
	c.data[k] = v
}

func (c *config) get(k string) interface{} {
	return c.data[k]
}

func convert(v reflect.Value, typ reflect.Type) reflect.Value {
	if v.Kind() == reflect.Interface {
		return convert(v.Elem(), typ)
	}
	switch typ.Kind() {
	case reflect.Ptr:
		res := reflect.New(typ.Elem())
		res.Elem().Set(convert(v, typ.Elem()))
		return res
	case reflect.Struct:
		if v.Kind() != reflect.Map {
			return v
		}
		res := reflect.New(typ).Elem()
		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)
			fieldVal := v.MapIndex(reflect.ValueOf(field.Name))
			if !fieldVal.IsValid() {
				fieldVal = v.MapIndex(reflect.ValueOf(strings.ToLower(field.Name)))
				if !fieldVal.IsValid() {
					continue
				}
			}
			res.Field(i).Set(convert(fieldVal, field.Type))
		}
		return res
	case reflect.Slice, reflect.Array:
		if v.Kind() != reflect.Slice {
			return v
		}
		res := reflect.Zero(typ)
		elemType := typ.Elem()
		for i := 0; i < v.Len(); i++ {
			res = reflect.Append(res, convert(v.Index(i), elemType))
		}
		return res
	default:
		if !v.Type().ConvertibleTo(typ) {
			return v
		}
		return v.Convert(typ)
	}
}
