// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

//go:generate hel

package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
	"gopkg.in/yaml.v2"
)

type unmarshaller func(io.Reader, interface{}) error

type marshaller func(io.Writer, interface{}) error

type codec struct {
	unmarshal unmarshaller
	marshal   marshaller
}

var (
	codecJSON = codec{
		unmarshal: func(r io.Reader, t interface{}) error {
			// json.Decoder seems to react very differently from other decoders, refusing to
			// actually read until EOF and instead stopping after a call to Read returns a
			// number less than the length of the byte slice passed in.  This is ... incorrect
			// behavior, to say the least ... so we're doing the reading manually.
			b, err := ioutil.ReadAll(r)
			if err != nil {
				return err
			}
			return json.Unmarshal(b, t)
		},
		marshal: func(w io.Writer, t interface{}) error {
			return json.NewEncoder(w).Encode(t)
		},
	}

	codecYAML = codec{
		unmarshal: func(r io.Reader, t interface{}) error {
			return yaml.NewDecoder(r).Decode(t)
		},
		marshal: func(w io.Writer, t interface{}) error {
			return yaml.NewEncoder(w).Encode(t)
		},
	}

	codecTOML = codec{
		unmarshal: func(r io.Reader, t interface{}) error {
			_, err := toml.DecodeReader(r, t)
			return err
		},
		marshal: func(w io.Writer, t interface{}) error {
			return toml.NewEncoder(w).Encode(t)
		},
	}
)

// Opener is a type which can read files.
type Opener interface {
	Open(path string) (f io.ReadCloser, err error)
	Create(path string) (f io.WriteCloser, err error)
}

// Config is a type to mirror viper.Config, but with better custom
// type support.  This stems from the closure of
// github.com/spf13/viper#271.
type Config struct {
	opener  Opener
	name    string
	dirs    []string
	formats []string
	codecs  map[string]codec

	writePath  string
	choseCodec codec

	// TODO: use an fsw.Watcher to watch for filesystem changes and
	// reload.
	data map[string]interface{}
}

// New returns a Config for the given Config file name (without
// extension) and list of valid directories for that file to be found in.
func New(opener Opener, name string, dirs ...string) (*Config, error) {
	c := &Config{
		opener: opener,
		name:   name,
		dirs:   dirs,
		formats: []string{
			"toml",
			"yaml",
			"yml",
			"json",
		},
		codecs: map[string]codec{
			"toml": codecTOML,
			"yaml": codecYAML,
			"yml":  codecYAML,
			"json": codecJSON,
		},
		data: make(map[string]interface{}),
	}
	if err := c.load(); err != nil && !os.IsNotExist(err) {
		return nil, err
	}
	return c, nil
}

func (c *Config) bestFile() (io.ReadCloser, unmarshaller, error) {
	if len(c.dirs) == 0 {
		return nil, nil, errors.New("cannot load a config file with no directories")
	}
	for _, d := range c.dirs {
		for _, ext := range c.formats {
			path := filepath.Join(d, fmt.Sprintf("%s.%s", c.name, ext))
			f, err := c.opener.Open(path)
			if os.IsNotExist(err) {
				continue
			}
			if err != nil {
				return nil, nil, err
			}
			c.writePath = path
			c.choseCodec = c.codecs[ext]
			return f, c.codecs[ext].unmarshal, nil
		}
	}
	c.writePath = filepath.Join(c.dirs[0], fmt.Sprintf("%s.%s", c.name, c.formats[0]))
	c.choseCodec = codecTOML
	return nil, nil, os.ErrNotExist
}

func (c *Config) load() error {
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
			return fmt.Errorf("duplicate keys %s and %s found; Configs are case insensitive", k, lk)
		}
		c.data[lk] = v
		delete(c.data, k)
	}
	return nil
}

// SetDefault sets the type and the default value for the data at k.
func (c *Config) SetDefault(k string, v interface{}) {
	k = strings.ToLower(k)
	if d, ok := c.data[k]; ok {
		c.data[k] = convert(reflect.ValueOf(d), reflect.TypeOf(v)).Interface()
		return
	}
	c.Set(k, v)
}

// Set sets the value at k.
func (c *Config) Set(k string, v interface{}) {
	k = strings.ToLower(k)
	c.data[k] = v
}

// Get gets the value at k.  If the default value has been set using SetDefault,
// the data will be converted to the same type as the default value.
func (c *Config) Get(k string) interface{} {
	return c.data[k]
}

// Keys returns a list of all keys available in c.
func (c *Config) Keys() []string {
	var l []string
	for k := range c.data {
		l = append(l, k)
	}
	return l
}

// Write writes c to the path it was opened from, or the most preferred path
// path otherwise.
func (c *Config) Write() error {
	f, err := c.opener.Create(c.writePath)
	if err != nil {
		return fmt.Errorf("could not create config file %s: %s", c.writePath, err)
	}
	defer f.Close()
	if err := c.choseCodec.marshal(f, c.data); err != nil {
		return fmt.Errorf("could not marshal settings: %s", err)
	}
	return nil
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
	case reflect.Map:
		if v.Kind() != reflect.Map {
			return v
		}
		res := reflect.MakeMap(typ)
		keyType := typ.Key()
		elemType := typ.Elem()
		iter := v.MapRange()
		for iter.Next() {
			res.SetMapIndex(convert(iter.Key(), keyType), convert(iter.Value(), elemType))
		}
		return res
	default:
		if !v.Type().ConvertibleTo(typ) {
			return v
		}
		return v.Convert(typ)
	}
}
