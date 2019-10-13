// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"go/token"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"

	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/fsw"
	"github.com/nelsam/vidar/setting"
)

var (
	dirColor = gxui.Color{
		R: 0.8,
		G: 1,
		B: 0.7,
		A: 1,
	}
	dropColor = gxui.Color{
		R: 0.7,
		G: 0.7,
		B: 0.1,
		A: 1,
	}
	fileColor = gxui.Gray80

	splitterBarBackgroundColor = gxui.Color{
		R: 0.6,
		G: 0.3,
		B: 0.3,
		A: 1,
	}
	splitterBarForegroundColor = gxui.Color{
		R: 0.8,
		G: 0.4,
		B: 0.4,
		A: 1,
	}
)

type Locationer interface {
	File() string
	Position() token.Position
}

type ProjectTree struct {
	button gxui.Button

	cmdr   Commander
	driver gxui.Driver
	theme  *basic.Theme

	dirs    *directory
	tocCtl  gxui.Control
	toc     *TOC
	tocLock sync.RWMutex

	watcher    fsw.Watcher
	reloadLock chan struct{}

	layout *splitterLayout
}

func NewProjectTree(cmdr Commander, driver gxui.Driver, window gxui.Window, theme *basic.Theme) *ProjectTree {
	tree := &ProjectTree{
		cmdr:       cmdr,
		driver:     driver,
		theme:      theme,
		reloadLock: make(chan struct{}, 1),
		button:     createIconButton(driver, theme, "folder.png"),
		layout:     newSplitterLayout(window, theme),
	}
	tree.initWatcher()
	tree.layout.SetOrientation(gxui.Vertical)
	tree.SetRoot(setting.DefaultProject.Path)

	return tree
}

func (p *ProjectTree) initWatcher() {
	w, err := fsw.New()
	if err != nil {
		// TODO: report to the UI
		log.Printf("WARNING: could not watch project tree: %s", err)
		return
	}
	p.watcher = w
	go p.watch()
}

func (p *ProjectTree) SetTOC(toc *TOC) {
	p.tocLock.Lock()
	defer p.tocLock.Unlock()
	p.toc = toc
}

func (p *ProjectTree) TOC() *TOC {
	p.tocLock.RLock()
	defer p.tocLock.RUnlock()
	return p.toc
}

func (p *ProjectTree) Button() gxui.Button {
	return p.button
}

func (p *ProjectTree) SetRoot(path string) {
	p.layout.RemoveAll()
	p.SetTOC(nil)
	p.tocCtl = nil

	if p.watcher != nil {
		if err := p.watcher.RemoveAll(); err != nil {
			log.Printf("WARNING: failed to remove current watches from watcher: %s", err)
		}
	}

	p.driver.Call(func() {
		p.dirs = newDirectory(p, path, p.watcher)
		scrollable := p.theme.CreateScrollLayout()
		// Disable horiz scrolling until we can figure out an accurate
		// way to calculate our width.
		scrollable.SetScrollAxis(false, true)
		scrollable.SetChild(p.dirs)
		p.layout.AddChild(scrollable)
		p.layout.SetChildWeight(p.dirs, 1)

		// Expand the top level
		p.dirs.button.Click(gxui.MouseEvent{})

		p.layout.Relayout()
		p.layout.Redraw()
	})
}

// watch waits for events from p.watcher.  For each event, the tree will
// spin off a goroutine to update the its children.
//
// Events are processed in separate goroutines to help us keep up with
// rapidly occurring events, e.g. in the case of a `git checkout` that
// touches many, many files and directories.  It doesn't completely
// prevent UI lock up, but it mitigates it some.
func (p *ProjectTree) watch() {
	for {
		e, err := p.watcher.Next()
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Printf("ProjectTree: Error from watcher: %s", err)
		}
		switch e.Op {
		case fsw.Write, fsw.Create, fsw.Remove, fsw.Rename:
			go p.update(e.Path)
		}
	}
}

// update will update any parts of p that changes to path would affect.
//
// If other concurrent calls to update are running, update may bail out
// in order to prevent locking up the UI.
//
// KNOWN ISSUE: This logic could cause the UI to get out of sync with
// the filesystem.  Either an event for /bar could be ignored while
// we process an event for /foo, or an event for /foo could be ignored
// when the current state of the filesystem has already been processed,
// but before the lock has been released.
//
// So far, though, I haven't seen a situation where fs events overwhelm
// this logic to the point that the UI displays incorrect data, so maybe
// it's good enough?  I'll be keeping my eye out for the UI getting in
// to a bad state, but I'm not going to solve the issue until I know
// it really is an issue.
func (p *ProjectTree) update(path string) {
	select {
	case p.reloadLock <- struct{}{}:
	default:
		return
	}
	defer func() {
		<-p.reloadLock
	}()

	p.driver.CallSync(func() {
		p.dirs.update(path)
	})
	toc := p.TOC()
	if toc != nil && strings.HasPrefix(path, toc.dir) {
		p.driver.CallSync(toc.Reload)
	}
}

func (p *ProjectTree) SetProject(project setting.Project) {
	// Ensure that the project tree is the current pane before
	// the UI goroutine does our relayout/redraw logic.
	p.driver.Call(func() {
		if p.layout.Attached() {
			return
		}
		p.button.Click(gxui.MouseEvent{
			Button: gxui.MouseButtonLeft,
		})
	})
	p.SetRoot(project.Path)
}

func (p *ProjectTree) Open(path string, pos token.Position) {
	dir, _ := filepath.Split(path)
	p.dirs.ExpandTo(dir)
}

func (p *ProjectTree) Frame() gxui.Control {
	return p.layout
}

type splitterLayout struct {
	mixins.SplitterLayout

	window gxui.Window
	theme  gxui.Theme
}

func newSplitterLayout(window gxui.Window, theme gxui.Theme) *splitterLayout {
	l := &splitterLayout{
		window: window,
		theme:  theme,
	}
	l.Init(l, theme)
	return l
}

func (l *splitterLayout) DesiredSize(min, max math.Size) math.Size {
	s := l.SplitterLayout.DesiredSize(min, max)
	width := 20 * l.theme.DefaultMonospaceFont().GlyphMaxSize().W
	if min.W > width {
		width = min.W
	}
	if max.W < width {
		width = max.W
	}
	s.W = width
	return s
}

func (l *splitterLayout) CreateSplitterBar() gxui.Control {
	bar := editor.NewSplitterBar(l.window.Viewport(), l.theme)
	bar.OnSplitterDragged(func(wndPnt math.Point) { l.SplitterDragged(bar, wndPnt) })
	return bar
}

type parent interface {
	Children() gxui.Children
}

type irrespParent interface {
	MissingChild() gxui.Control
}
