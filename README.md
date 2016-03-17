# vidar

[![Join the chat at https://gitter.im/nelsam/vidar](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/nelsam/vidar?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)

Vidar is the Norse god of silence, patience, and revenge.  Sounds perfect for an editor, right?

## History

Vidar started as a repository that I had named `gxui_playground`.  It was quite literally just a
place for me to mess around with [gxui](https://github.com/google/gxui).  The problem I chose to tackle
for my little learning exercise was "I dislike nearly every editor out there for working on Go
code."

Three weeks later, and I have something that I'm actually *starting* to use for development.  Sure,
it isn't terribly stable, has a high chance of data loss, has zero tests, and is still missing
tons of features that I would argue are necessary for a good text editor...

But hey, considering how long I've been working on it, I'm pretty happy.

## Goals

Mostly, I want my old emacs config, but with modern keybindings and a UI library that actually
works correctly cross-platform (I've had a lot of trouble getting mouse clicks in OS X to
register with emacs's UI).  I have hopes of taking lessons from those old editors and some new
ones.

If I can make something that I find useful for editing Go code, I'll be overjoyed.  If anyone
else finds it useful, I'll probably die of excitement.  Which means ownership transfers to the
person who found it useful (as they're indirectly responsible for my death).

## Currently

It's getting closer to a usable state.  I've removed most of the panics (now, panics are only
used in cases where I've screwed up completely), so the potential for data loss is low,
at this point.  It has copy/cut/paste and undo/redo support, and a lot of the weirder bugs
from gxui have been fixed in a fork I'm maintaining (gxui got abandoned).  Right now, I do
most of my vidar development using vidar.  The only things I don't yet use it for revolve
around find/replace stuff.

I still don't think it's a good replacement for your every day editor, even for Go code, but
it does have some features that I find lacking in the other editors out there.  If you'd like
to try it, be my guest.  I've added a file/edit menu (soon to have a view menu) which displays
commands and their key bindings, so you no longer have to learn key bindings from the source
code.

Code suggestions are still a little weird - gocode is doing something I don't expect with
regard to the potential suggestions in the current package.  Imported packages and the
current file both work, but other files in the current package ... don't.  I don't know
why.

## Requirements

If you do decide to try it, it should be noted that you'll have a better time if you have
gocode (for suggestions) and goimports (for formatting the file automatically on save) in
your $PATH.
