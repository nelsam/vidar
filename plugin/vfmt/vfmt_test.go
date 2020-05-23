// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package vfmt_test

import (
	"testing"

	"github.com/nelsam/hel/pers"
	"github.com/nelsam/vidar/command"
	"github.com/nelsam/vidar/commander/bind"
	"github.com/nelsam/vidar/commander/text"
	"github.com/nelsam/vidar/plugin/vfmt"
	"github.com/nelsam/vidar/setting"
	"github.com/nelsam/vidar/test"
	"github.com/poy/onpar"
	"github.com/poy/onpar/expect"
)

func TestOnSave(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	var (
		_ command.BeforeSaver = vfmt.OnSave{}
		_ bind.OpHook         = vfmt.OnSave{}
	)

	o.BeforeEach(func(t *testing.T) (expectation, vfmt.OnSave) {
		return expect.New(t, expect.WithDiffer(test.Differ())), vfmt.OnSave{}
	})

	o.Spec("it corrects v syntax", func(expect expectation, o vfmt.OnSave) {
		before := `
module main





fn do_stuff(){
println('stuff is done')
}

fn main(){
do_stuff()



and_other_stuff()
}


fn and_other_stuff() { println('some more stuff is done') }
`
		expected := `module main

fn do_stuff() {
	println('stuff is done')
}

fn main() {
	do_stuff()
	and_other_stuff()
}

fn and_other_stuff() {
	println('some more stuff is done')
}
`
		nt, err := o.BeforeSave(setting.Project{}, "/path/to/a/file.v", before)
		expect(err).To(not(haveOccurred()))
		expect(nt).To(equal(expected))
	})
}

func TestVfmt(t *testing.T) {
	o := onpar.New()
	defer o.Run(t)

	var _ bind.Command = &vfmt.Vfmt{}

	o.BeforeEach(func(t *testing.T) (expectation, *vfmt.Vfmt, *mockEditor, *mockApplier) {
		return expect.New(t, expect.WithDiffer(test.Differ())), vfmt.New(newMockLabelCreator()), newMockEditor(), newMockApplier()
	})

	o.Spec("it uses the editor's text as the old", func(expect expectation, v *vfmt.Vfmt, ed *mockEditor, a *mockApplier) {
		source := `
module main


fn foo() {
	println('foo')
}`
		formatted := `module main

fn foo() {
	println('foo')
}
`
		pers.Return(ed.TextOutput, source)
		expect(v.Store(ed)).To(equal(bind.Waiting))
		expect(v.Store(a)).To(equal(bind.Done))
		expect(v.Exec()).To(not(haveOccurred()))
		expect(a).To(haveMethodExecuted("Apply", withArgs(ed, text.Edit{
			At:  0,
			Old: []rune(source),
			New: []rune(formatted),
		})))
	})
}
