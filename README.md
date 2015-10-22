# vidar

Vidar is the Norse god of silence, patience, and revenge.  Sounds perfect for an editor, right?

## History

Vidar started as a repository that I had named `gxui_playground`.  It was quite literally just a
place for me to mess around with [gxui](github.com/google/gxui).  The problem I chose to tackle
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

Don't use it.  Really.  It does compile, and it does technically open files.  Just don't rely
on it *staying* open.  I'm using panics to let me know when my syntax highlighter hits a part
of the language that it doesn't know how to highlight yet.  I'm not responsible if that happens
when you have unsaved changes.

I'm gradually using it more and more often to develop it, but my use case combines "I want to
edit code efficiently" with "I want to know which features I should be focusing on".  For the
second part, using the editor is the best way to know which feature it is most in need of.
