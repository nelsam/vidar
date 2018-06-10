// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package ui

// Wrapper represents any type that's wrapping a more complicated
// type from an upstream UI library.  You should always try to use
// the methods provided on the type rather than accessing the
// parent directly.
type Wrapper interface {
	// Parent returns the raw underlying type from the parent
	// library.  Ideally, you'll never need to use this.  If you
	// do need to use it, please create a github issue so that we
	// can add the needed functionality to the wrapper type.
	//
	// You must type assert the returned value to the type you
	// expect it to be.
	Parent() interface{}
}
