// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package lsp

import (
	"github.com/nelsam/vidar/theme"
	"github.com/sourcegraph/go-lsp"
)

func Construct(k lsp.SymbolKind) theme.LanguageConstruct {
	switch k {
	case lsp.SKKey:
		return theme.Keyword
	case lsp.SKFunction, lsp.SKMethod, lsp.SKConstructor:
		return theme.Func
	case lsp.SKClass, lsp.SKInterface:
		return theme.Type
	case lsp.SKString:
		return theme.String
	case lsp.SKNumber:
		return theme.Num
	case lsp.SKNull:
		return theme.Nil
	}
	return -1
}
