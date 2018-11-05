Go Syntax Highlighting!
-----------------------

The gosyntax plugin adds in syntax highlighting for `*.go` files.

## Syntax Assumptions

Much of the syntax highlighting assumes gofmted code.  If your code is not formatted that way,
some of the highlighting may be off.  For example, if you use `}else{` instead of `} else {`,
the wrong characters will be highlighted.  This should only have an affect on those particular
instances, though, and the highlighting for the rest of the file should be fine.
