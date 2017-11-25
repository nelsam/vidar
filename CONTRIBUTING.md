Contributing to Vidar
---------------------

Thanks for thinking about contributing!

## Quick Disclaimer

Vidar grew up as a pet project that I never expected to gain traction.  It's lacking in tests
and I've taken some shortcuts that frustrate me, and I expect them to frustrate you too.  I
am certainly not opposed to accepting drastic, aggressive changes as long as they make the
system better.  However, also be aware that *some* changes (specifically, anything that breaks
backward compatibility) may result in some requested changes.

## Plugins

Want to add features that might not be desired by every user?  Consider building a
[plugin](https://github.com/nelsam/vidar/wiki/Plugin-Development) and making a PR to
[the unofficial list on the wiki](https://github.com/nelsam/vidar/wiki/Unofficial-Plugins).

## Code Quality

I very closely follow the [Effective Go](https://golang.org/doc/effective_go.html) styles.
Please strive to follow them in your PRs, as well.

### Else Statements

In addition, you'll notice that there are very, very few `else` statements in the code. This
is to reduce branching logic as much as possible.  Please think about your code before you
submit a PR with else statements, but don't be afraid to submit it if you think it fits.
Just be aware that I will probably suggest an alternative.

### Tooling

I expect all code to pass:

* goimports
* go vet

I expect *most* code to pass [gometalinter](https://github.com/alecthomas/gometalinter).  There
are a number of exceptions, here, but give it your best effort.

Here's a quick summary of what I'll probably ignore from the gometalinter output:

* `should have comment or be unexported` errors for functions or methods that are completely
  obvious.  If the function name, parameter type(s), and return type(s) make it clear, I don't
  really care about adding documentation comments.
* `gocyclo` for anything using a `switch` statement.
  * `switch` statements in `go` are pretty damn clear, and don't deserve an extra point for
    every single case, from my perspective.
* `vetshadow`, in general.  Shadowing variables isn't really something I'm that worried about.
