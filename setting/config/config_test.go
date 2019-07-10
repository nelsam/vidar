// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package config_test

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"testing"
	"time"

	"github.com/nelsam/hel/pers"
	"github.com/nelsam/vidar/setting/config"
	"github.com/poy/onpar"
	"github.com/poy/onpar/expect"
	"github.com/poy/onpar/matchers"
)

type Expectation = expect.Expectation

var (
	Not          = matchers.Not
	HaveOccurred = matchers.HaveOccurred
	Equal        = matchers.Equal
	ViaPolling   = matchers.ViaPolling
	StartWith    = matchers.StartWith
)

func TestConfig(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	o.BeforeEach(func(t *testing.T) (Expectation, *mockOpener) {
		return expect.New(t), newMockOpener()
	})

	o.Spec("it propagates unexpected errors", func(expect Expectation, o *mockOpener) {
	})

	o.Spec("it works when there is no existing config file", func(expect Expectation, o *mockOpener) {
		done, err := pers.ConsistentlyReturn(o.OpenOutput, nil, os.ErrNotExist)
		expect(err).To(Not(HaveOccurred()))
		defer done()

		c, err := config.New(o, "foo", "/bar", "/baz")
		expect(err).To(Not(HaveOccurred()))
		expect(c.Get("foo")).To(Equal(nil))
		c.Set("foo", "bar")
		expect(c.Get("foo")).To(Equal("bar"))
	})

	type ret struct {
		c   *config.Config
		err error
	}

	newConfig := func(expect Expectation, o *mockOpener, fpath, contents, name string, dirs ...string) ret {
		r := make(chan ret)
		go func() {
			c, err := config.New(o, name, dirs...)
			r <- ret{
				c:   c,
				err: err,
			}
		}()
		expect(func() string {
			select {
			case path := <-o.OpenInput.Path:
				if path != fpath {
					pers.Return(o.OpenOutput, nil, os.ErrNotExist)
				}
				return path
			case <-time.After(100 * time.Millisecond):
				return ""
			}
		}).To(ViaPolling(Equal(fpath)))

		f := newMockReadCloser()
		close(f.CloseOutput.Ret0)
		expect(pers.Return(o.OpenOutput, f, nil)).To(Not(HaveOccurred()))
		b := bytes.NewBufferString(contents)
		expect(func() (done bool) {
			select {
			case p := <-f.ReadInput.P:
				n, err := b.Read(p)
				f.ReadOutput.N <- n
				f.ReadOutput.Err <- err
				return err == io.EOF
			case <-time.After(100 * time.Millisecond):
				return false
			}
		}).To(ViaPolling(Equal(true)))
		return <-r
	}

	typTests := []struct {
		typ  string
		body string
	}{
		{typ: "toml", body: `bacon = "eggs"`},
		{typ: "yaml", body: `bacon: eggs`},
		{typ: "yml", body: `bacon: eggs`},
		{typ: "json", body: `{"bacon": "eggs"}`},
	}
	for i := range typTests {
		// We need to assign here to avoid having all the tests use the same memory location
		// for tt, which would cause them to all test json when run in parallel.
		tt := typTests[i]
		o.Spec(fmt.Sprintf("it can load %s files", tt.typ), func(expect Expectation, o *mockOpener) {
			ret := newConfig(expect, o, fmt.Sprintf("/bar/foo.%s", tt.typ), tt.body, "foo", "/bar")
			expect(ret.err).To(Not(HaveOccurred()))
			expect(ret.c.Get("bacon")).To(Equal("eggs"))
		})
	}

	o.Spec("it loads from primary paths", func(expect Expectation, o *mockOpener) {
		ret := newConfig(expect, o, "/bar/foo.toml", `foo = "bar"`, "foo", "/bar", "/baz")
		expect(ret.err).To(Not(HaveOccurred()))
		expect(ret.c.Get("foo")).To(Equal("bar"))
	})

	o.Spec("it loads from secondary paths", func(expect Expectation, o *mockOpener) {
		ret := newConfig(expect, o, "/baz/foo.toml", `foo = "bar"`, "foo", "/bar", "/baz")
		expect(ret.err).To(Not(HaveOccurred()))
		expect(ret.c.Get("foo")).To(Equal("bar"))
	})
}
