// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax

import (
	"unicode"

	"github.com/nelsam/vidar/commander/input"
	"github.com/nelsam/vidar/theme"
)

type scopeMap struct {
	start, end int
	typ        Scope
	construct  theme.LanguageConstruct
	constructs []input.SyntaxLayer
	nested     []*scopeMap
	parent     *scopeMap
}

func (m scopeMap) depth(pos int) int {
	if pos < m.start || pos > m.end {
		return -1
	}
	for _, c := range m.constructs {
		if c.Contains(pos) {
			return 0
		}
	}
	for _, n := range m.nested {
		if d := n.depth(pos); d != -1 {
			return d + 1
		}
	}
	return 0
}

func (m scopeMap) layers() []input.SyntaxLayer {
	scope := []input.SyntaxLayer{
		{Construct: m.construct, Spans: []input.Span{
			{Start: m.start, End: m.start + len(m.typ.Open)},
			{Start: m.end, End: m.end + len(m.typ.Close)},
		}},
	}
	layers := append(scope, m.constructs...)
	for _, n := range m.nested {
		layers = append(layers, n.layers()...)
	}
	return layers
}

// Map is a representation of the mapped syntax of a file.  It knows
// about scopes and various language constructs.
type Map struct {
	file scopeMap
}

// Layers returns all syntax layers from all nested scopes.
func (m Map) Layers() []input.SyntaxLayer {
	return m.file.layers()
}

// Depth returns the scope depth at pos.
func (m Map) Depth(pos int) int {
	return m.file.depth(pos)
}

// Scope represents an opening and closing scope.  These will
// always use the theme.ScopePair language construct.
type Scope struct {
	// Open is the string that opens a scope.  For example, { is a
	// common way to open a scope.
	Open string

	// Close is the string that closes a scope.  For example, } is
	// a common way to close a scope.
	Close string
}

// Wrapped represents a wrapped type, like a string or comment block.
type Wrapped struct {
	// Open is the string that opens a the wrapped type, e.g. " or '.
	Open string

	// Close is the string that closes the wrapped type.  Can be the
	// same as Open.
	Close string

	// Escapes is any way that one could escape a Close.
	Escapes []string

	// Construct is the language construct that this wrapped type
	// detects.
	Construct theme.LanguageConstruct

	// Nested tells us whether or not this type can nest itself
	Nested bool
}

func (w Wrapped) skip(d []rune) int {
	for _, e := range w.Escapes {
		esc := []rune(e)
		if len(d) >= len(esc) && match(esc, d[:len(esc)]) {
			return len(esc)
		}
	}
	if w.Nested {
		open := []rune(w.Open)
		if len(d) >= len(open) && match(open, d[:len(open)]) {
			return w.end(d[len(open):])
		}
	}
	c := []rune(w.Close)
	if len(d) >= len(c) && match(c, d[:len(c)]) {
		return 0
	}
	return 1
}

func (w Wrapped) end(d []rune) int {
	for i := 0; i < len(d); {
		skip := w.skip(d[i:])
		if skip == 0 {
			return i + len(w.Close)
		}
		i += skip
	}
	return len(d)
}

// Word represents a word surrounded by defining characteristics
// which should be a certain type of language construct.
//
// A simple example: a go function can be detected with:
//     syntax.Word{Before: []rune("func"), Construct: theme.Func}
//
// TODO: for many languages, using words before and after is not
// sufficient; but for now, because we don't want to mess with
// the problems of multiple spaces, we're just going to leave it.
type Word struct {
	// Before is the word (without spaces) that prefixes this
	// construct.  Do not list separators (e.g. } or )).
	Before string

	// After is the word (without spaces) that suffixes this
	// construct.  Do not list separators (e.g. { or ().
	After string

	// Construct is the language construct that this word detects.
	Construct theme.LanguageConstruct
}

// Generic understands and parses a general map of a language syntax.
type Generic struct {
	Scopes       []Scope
	Wrapped      []Wrapped
	StaticWords  map[string]theme.LanguageConstruct
	DynamicWords []Word
}

func (g Generic) matchScope(d []rune) *Scope {
	for _, s := range g.Scopes {
		open := []rune(s.Open)
		if len(d) >= len(open) && match(open, d[:len(open)]) {
			return &s
		}
	}
	return nil
}

func (g Generic) matchWrapped(d []rune) *Wrapped {
	for _, w := range g.Wrapped {
		open := []rune(w.Open)
		if len(d) >= len(open) && match(open, d[:len(open)]) {
			return &w
		}
	}
	return nil
}

func (g Generic) dynWord(before, after []rune) (theme.LanguageConstruct, bool) {
	for _, w := range g.DynamicWords {
		mBefore := []rune(w.Before)
		if len(mBefore) > 0 && !match(before, mBefore) {
			continue
		}
		mAfter := []rune(w.After)
		if len(mAfter) > 0 && !match(after, mAfter) {
			continue
		}
		return w.Construct, true
	}
	return 0, false
}

func (g Generic) Parse(d []rune) Map {
	curr := &scopeMap{}
	rainbow := theme.ScopePair
	var lastWord []rune
	for i := 0; i < len(d); {
		remaining := d[i:]
		end := []rune(curr.typ.Close)
		if len(end) > 0 && len(remaining) >= len(end) && match(end, remaining[:len(end)]) {
			rainbow--
			curr.end = i
			curr = curr.parent
			i += len(end)
			continue
		}
		if s := g.matchScope(remaining); s != nil {
			n := &scopeMap{start: i, construct: rainbow, typ: *s, parent: curr}
			rainbow++
			curr.nested = append(curr.nested, n)
			curr = n
			i += len(s.Open)
			continue
		}
		if w := g.matchWrapped(remaining); w != nil {
			remaining = remaining[len(w.Open):]
			length := len(w.Open) + w.end(remaining)
			curr.constructs = append(curr.constructs, input.SyntaxLayer{
				Construct: w.Construct,
				Spans:     []input.Span{{Start: i, End: i + length}},
			})
			i += length
			continue
		}

		word := nextWord(remaining)
		if len(word) == 0 {
			i++
			continue
		}
		wordStart := i
		i += len(word)

		if isNumeric(word) {
			curr.constructs = append(curr.constructs, input.SyntaxLayer{
				Construct: theme.Num,
				Spans:     []input.Span{{Start: wordStart, End: i}},
			})
			lastWord = word
			continue
		}
		c, ok := g.StaticWords[string(word)]
		if !ok {
			j := i
			for ; j < len(d) && isSeparator(d[j]); j++ {
			}
			next := nextWord(d[j:])
			c, ok = g.dynWord(lastWord, next)
		}
		if ok {
			curr.constructs = append(curr.constructs, input.SyntaxLayer{
				Construct: c,
				Spans:     []input.Span{{Start: wordStart, End: i}},
			})
			lastWord = word
			continue
		}
	}
	if curr.end == 0 {
		curr.end = len(d)
	}
	for curr.parent != nil {
		curr = curr.parent
	}
	return Map{file: *curr}
}

func nextWord(v []rune) (word []rune) {
	for _, v := range v {
		if isSeparator(v) {
			break
		}
		word = append(word, v)
	}
	return word
}

func isSeparator(r rune) bool {
	// TODO: the separator characters need to be customizable.
	switch r {
	case ' ', '\t', '\n', ';', '(', ')', '{', '}', ':', '[', ']', ',':
		return true
	default:
		return false
	}
}

func isNumeric(w []rune) bool {
	decimal := false
	for _, r := range w {
		if r == '.' && !decimal {
			decimal = true // there can only be one
			continue
		}
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
func match(a, b []rune) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
