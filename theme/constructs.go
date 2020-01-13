// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package theme

type LanguageConstruct int

const (
	Keyword LanguageConstruct = 1 + iota
	Builtin
	Func
	Type
	Ident

	String
	Num
	Nil

	Comment

	Bad

	// ScopePair is a much higher value to provide extra space
	// for other language constructs (e.g. for languages that
	// have constructs that Go doesn't).  Because ScopePairs are
	// often intentionally highlighted with rainbow so that each
	// pair of opening/closing marks are a different color,
	// ScopePair can be used as an initial value, with the value
	// being incremented for each nested pair.
	//
	// For example, the opening/closing marks for functions may
	// be ScopePair, while opening/closing marks for if statements
	// at the top level of functions may be ScopePair+1.
	ScopePair LanguageConstruct = 100
)
