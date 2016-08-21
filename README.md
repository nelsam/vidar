[![Build Status](https://travis-ci.org/nelsam/vidar.svg?branch=master)](https://travis-ci.org/nelsam/vidar)

# vidar

[![Join the chat at https://gitter.im/nelsam/vidar](https://badges.gitter.im/Join%20Chat.svg)](https://gitter.im/nelsam/vidar?utm_source=badge&utm_medium=badge&utm_campaign=pr-badge&utm_content=badge)
[![Issues Ready](https://badge.waffle.io/nelsam/vidar.svg?label=ready&title=Ready)](http://waffle.io/nelsam/vidar)
[![Go Report Card](https://goreportcard.com/badge/github.com/nelsam/vidar)](https://goreportcard.com/report/github.com/nelsam/vidar)

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

## Goals

Mostly, I want my old emacs config, but with modern keybindings and a UI library that actually
works correctly cross-platform (I've had a lot of trouble getting mouse clicks in OS X to
register with emacs's UI).  I have hopes of taking lessons from those old editors and some new
ones.

If I can make something that I find useful for editing Go code, I'll be overjoyed.  If anyone
else finds it useful, I'll probably die of excitement.  Which means ownership transfers to the
person who found it useful (as they're indirectly responsible for my death).

## Currently

It's in a fairly usable state.  The potential for data loss is low.  It supports most of what
you'd expect from a basic text editor, there are some extra features thrown in, and I'm
maintaining some fixes for gxui bugs in a fork (I plan to make PRs when I can, but upstream
is mostly unmaintained, so it's not my highest priority at the moment).

I doubt it will replace your favorite text editor or IDE in its current state, but if you're
curious or just unhappy with all of the Go editors out there, feel free to give it a shot.
I welcome issues and pull requests.

## Requirements

If you do decide to try it, it should be noted that you'll have a better time if you have
gocode (for suggestions), godef (for go-to-definition), and goimports (for formatting the
file automatically on save) in your $PATH.  It should work without them, but it'll work a
lot nicer with them.

## Installation

`go get github.com/nelsam/vidar`

I don't believe in vendoring source code, so this may break from time to time.  I plan
to have a build system that builds `.deb`, `.rpm`, `.dmg`, and `.exe` files in CI, but
it's not yet in place.

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
