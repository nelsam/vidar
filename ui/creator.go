// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package ui

import "github.com/nelsam/vidar/theme"

// Creator represents a type which can create UI elements.
type Creator interface {
	// Name is the name of the UI library, for use in user
	// interactions. For example, the user may want to bind a
	// command line flag to the use of a curses-based ui.Creator,
	// for use with the EDITOR environment variable.
	Name() string

	// Start initializes the UI library and kicks off any
	// goroutines the library needs running in order to operate.
	Start() error

	// Wait waits for the last window to be closed.
	Wait()

	// Quit closes all windows and stops all goroutines.
	Quit()

	// Theme returns the theme of the creator.
	Theme() theme.Theme

	// Runner returns the type that knows how to run logic on the UI
	// goroutine.
	Runner() Runner

	// Window creates a window in the UI.
	Window(size Size, name string) (Window, error)

	// LinearLayout creates a layout which adds elements linearly.
	// start must be a single direction.  The first child will
	// be added at that side of the element, with each subsequent
	// child being added to the opposite side of the previous
	// child.
	LinearLayout(start Direction) (Layout, error)

	// ManualLayout creates a layout which adds elements manually.
	// The caller should expect elements to be added at the top
	// left of the layout, and are expected to use LayoutOpts to
	// arrange the added children correctly.
	ManualLayout() (Layout, error)

	// TabbedLayout creates a layout which adds elements in new
	// tabs.  All tabs are expected to have a name passed in as
	// a layout option, and may error if one isn't provided.
	TabbedLayout() (Layout, error)

	// SplitLayout creates a layout which splits available space
	// between its children, separating children with a bar.
	SplitLayout() (Layout, error)

	// Label creates a label in the UI with the passed in string
	// as its text.
	Label([]rune) (TextDisplay, error)

	// Editor creates a text editor element in the UI.
	Editor() (TextEditor, error)
}
