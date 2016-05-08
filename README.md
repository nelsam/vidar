# vidar

[![Join the chat at https://gitter.im/nelsam/vidar](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/nelsam/vidar?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge) [![Issues Ready](https://badge.waffle.io/nelsam/vidar.svg?label=ready&title=Ready)](http://waffle.io/nelsam/vidar)

Vidar is the Norse god of silence, patience, and revenge.  Sounds perfect for an editor, right?

## History

Vidar started as a repository that I had named `gxui_playground`.  It was quite literally just a
place for me to mess around with [gxui](https://github.com/google/gxui).  The problem I chose to tackle
for my little learning exercise was "I dislike nearly every editor out there for working on Go
code."  Three weeks later, I had something that I was actually *starting* to use for development.

I'm six months in, now, and things are shaping up pretty nicely.  I've solved some of the issues
that originated in GXUI, fixed almost all of the syntax highlighting issues, and have most of the
basic functionality of a modern editor.  If you want to see what is yet to be done, feel free
to check out the issues.

## Current Features

- Completion suggestions (via gocode)
- Go to definition (via godef)
- Split view (both horizontal and vertical)
- Watch filesystem for changes
  - Events trigger editor elements to reload their text
  - Since I don't trust this to be fully reliable, vidar will refuse to write a file that has
    been changed on disk since the last reload; but this doesn't happen often.
- Manage the license header of files (defined in a `.license-header` file in the project root)
  - If you're like me and always forget to add the license blurb at the top of every Go file,
    this is for you.
  - New files will be opened with the license blurb as the initial text.
  - There is a command (`ctrl+shift+l` or `cmd+shift+l`, currently) which will update the license
    blurb of the current file with the text from the `.license-header` file.
- Syntax highlighting
- Rainbow parens
- Most of the basic stuff you expect from a text editor (copy/paste, undo/redo, etc)

## Important Missing Features

These are all planned, but have yet to be implemented.

- Configurability
  - When you open a new project, you can set up some basic configuration, but other than that...
  - For example: key bindings are hardcoded, project settings cannot be changed after adding
    the project unless you edit the config file, you can't edit the font or theme, etc.
- Polish
  - There are some frustrating, but difficult-to-solve, bugs lingering around.  I squash them
    when I can, but some of the less annoying ones that either have difficult solutions or are
    difficult to reproduce regularly are going to be there for a while.
- Support for Plugins (including languages other than Go)
  - I've built the editor with plugins in mind, but it's not quite ready to have all that
    abstracted away, right now.  It'll get there.

## Goals

Mostly, I want my old emacs config, but with modern keybindings and a UI library that actually
works correctly cross-platform (I've had a lot of trouble getting mouse clicks in OS X to
register with emacs's UI).  I have hopes of taking lessons from those old editors and some new
ones.

If I can make something that I find useful for editing Go code, I'll be overjoyed.  If anyone
else finds it useful, I'll probably die of excitement.  Which means ownership transfers to the
person who found it useful (as they're indirectly responsible for my death).

## Currently

It's getting closer to a usable state.  The potential for data loss is low.  It supports most
of what you'd expect from a basic text editor, there are some extra features thrown in, and
I'm maintaining some fixes for gxui bugs in a fork (I plan to make PRs when I can, but upstream
is unmaintained, so it's not my highest priority at the moment).

The biggest blocker, from my perspective, is the fact that errors are mostly silent.  I haven't
yet hooked up an error or warning UI, so the only notice you get is from STDERR output.

If that's not a huge issue for you, feel free to give it a shot.  I welcome issues and pull
requests.

## Requirements

If you do decide to try it, it should be noted that you'll have a better time if you have
gocode (for suggestions), godef (for go-to-definition), and goimports (for formatting the
file automatically on save) in your $PATH.  It should work without them, but it'll work a
lot nicer with them.
