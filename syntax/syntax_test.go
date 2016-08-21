// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package syntax_test

import "testing"

func TestLayers(t *testing.T) {
	t.Run("Decl", Decl)
	t.Run("Unicode", Unicode)
	t.Run("PackageDocs", PackageDocs)
}
