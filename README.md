# vidar

[![Build Status](https://travis-ci.org/nelsam/vidar.svg?branch=master)](https://travis-ci.org/nelsam/vidar)
[![Join the chat at https://gitter.im/nelsam/vidar](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/nelsam/vidar?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Issues Ready](https://badge.waffle.io/nelsam/vidar.svg?label=ready&title=Ready)](http://waffle.io/nelsam/vidar)
[![Go Report Card](https://goreportcard.com/badge/github.com/nelsam/vidar)](https://goreportcard.com/report/github.com/nelsam/vidar)

Vidar is the Norse god of silence, patience, and revenge.  Sounds perfect for an editor, right?

## History

Vidar started as a repository that I had named `gxui_playground`.  It was quite literally just a
place for me to mess around with [gxui](https://github.com/google/gxui).  The problem I chose to tackle
for my little learning exercise was "I dislike nearly every editor out there for working on Go
code."  Three weeks later, I had something that I was actually *starting* to use for development.

Vidar is not a fully featured editor, right now.  However, it works for most purposes when editing
Go code, and I don't think it's far off from being in a finished 1.0 state.

## Installing

1. [Install GXUI's C dependencies](https://github.com/google/gxui#dependencies)
2. `go get github.com/nelsam/vidar`.

### Go Version

I'm only supporting the latest stable version of Go.  This doesn't necessarily mean that vidar
won't work with older versions, but I promise nothing.

### Optional Dependencies

- [gocode](https://github.com/nsf/gocode) for code suggestions
- [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports) for formatting files on save
  - This will some day be configurable, but it currently is not
- [godef](https://github.com/rogpeppe/godef) for goto-definition

### Binary Distribution

In the (hopefully near) future, I plan to have CI build installers and upload them to github
releases.  For the moment, though, I'm only distributing via source code.

## Configuration

Vidar uses [xdg-go](https://github.com/casimir/xdg-go) to decide where to save config
files.  On linux systems, this will probably end up in `~/.config/vidar/`; for Windows
and OS X, you'll likely need to check the xdg-go package to see what it uses.

Config files are written as `toml` by default, but can be parsed from any format that
[viper](https://github.com/spf13/viper) supports.  Currently, there are three config
files:
- settings: Only used to configure a `fonts` list, which should be a list of names
  of fonts installed on your system in order of preference.  Note that only truetype
  fonts are supported right now, and many of those display incorrectly.  My current
  favorites are `Inconsolata-Regular` and `PTM55F`.
- projects: A list of projects with `name`, `path`, and `gopath` keys.  This can be
  added to with the `add-project` command (`ctrl-shift-n` by default).
- keys: The key bindings.  This file will be written on first startup with the default
  key bindings, so you can edit the file with any changes or aliases you'd like.
  Multiple bindings per command are supported.

## Syntax Assumptions

Much of the syntax highlighting assumes gofmted code.  If your code is not formatted that way,
some of the highlighting may be off.  For example, if you use `}else{` instead of `} else {`,
the wrong characters will be highlighted.  This should only have an affect on those particular
instances, though, and the highlighting for the rest of the file should be fine.

## Current Features

- Completion suggestions (via gocode)
- Go to definition (via godef)
- Split view (both horizontal and vertical)
- Watch filesystem for changes
  - Events trigger editor elements to reload their text
  - Since this has shown itself to be a bit unreliable, vidar will refuse to write a file that
    has been changed on disk since the last reload; but this doesn't happen often.
- Manage the license header of files (defined in a `.license-header` file in the project root)
  - If you're like me and always forget to add the license blurb at the top of every Go file,
    this is for you.
  - New files will be opened with the license blurb as the initial text.
  - There is a command (`ctrl+shift+l` or `cmd+shift+l` by default) which will update the
    license blurb of the current file with the text from the `.license-header` file.
- Syntax highlighting
- Rainbow parens
- Most of the basic stuff you expect from a text editor (copy/paste, undo/redo, etc)

## Important Missing Features

These are all planned, but have yet to be implemented.

- Configurability
  - When you open a new project, you can set up some basic configuration in the UI, but very
    little else.
  - Most settings will require editing a config file manually, at the moment.
- Polish
  - There are some frustrating, but difficult-to-solve, bugs lingering around.  I squash them
    when I can, but some of the less annoying ones that either have difficult solutions or are
    difficult to reproduce regularly are going to be there for a while.
- Support for Plugins (including languages other than Go)
  - I've built the editor with plugins in mind, but it's not quite ready to have all that
    abstracted away, right now.  It'll get there.

