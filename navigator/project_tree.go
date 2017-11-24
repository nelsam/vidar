// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package navigator

import (
	"fmt"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/nelsam/gxui"
	"github.com/nelsam/gxui/math"
	"github.com/nelsam/gxui/mixins"
	"github.com/nelsam/gxui/themes/basic"
	"github.com/nelsam/vidar/editor"
	"github.com/nelsam/vidar/setting"
)

// maxWatchDirs is here just to ensure that we don't take too long to
// open up a project.  If the project has an excessive number of files,
// watching will fail when it hits this number.
//
// TODO: think about alternate limits, and possibly rely on polling
// when watching fails.
const maxWatchDirs = 4096

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

	watcher    *fsnotify.Watcher
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
	tree.layout.SetOrientation(gxui.Vertical)
	tree.SetProject(settings.DefaultProject)

	return tree
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
	defer p.driver.Call(func() {
		p.layout.Relayout()
		p.layout.Redraw()
	})
	p.layout.RemoveAll()
	p.SetTOC(nil)
	p.tocCtl = nil

	if p.watcher != nil {
		p.watcher.Close()
	}
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Printf("Error creating project tree watcher: %s", err)
	}
	p.watcher = watcher
	p.startWatch(path)

	p.dirs = newDirectory(p, path)
	scrollable := p.theme.CreateScrollLayout()
	// Disable horiz scrolling until we can figure out an accurate
	// way to calculate our width.
	scrollable.SetScrollAxis(false, true)
	scrollable.SetChild(p.dirs)
	p.layout.AddChild(scrollable)
	p.layout.SetChildWeight(p.dirs, 1)

	// Expand the top level
	p.dirs.button.Click(gxui.MouseEvent{})
}

func (p *ProjectTree) startWatch(root string) {
	if p.watcher == nil {
		return
	}

	count := 0
	err := filepath.Walk(root, func(path string, finfo os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("Error walking directory %s: %s", root, err)
		}
		if finfo.IsDir() {
			if filepath.Base(path)[0] == '.' {
				// This is mostly to skip .git directories, which contain
				// a pretty deep structure and can eat up a lot of our
				// watches.  Especially important on OS X, where the default
				// max number of watches is 256.
				return filepath.SkipDir
			}
			count++
			p.watcher.Add(path)
		}
		if count > maxWatchDirs {
			p.watcher.Close()
			return fmt.Errorf("Could not watch project: exceeded directory watch limit of %d", maxWatchDirs)
		}
		return nil
	})
	if err != nil {
		log.Printf("Warning: %s", err)
		return
	}
	go p.watch()
	go p.watchErrs()
}

func (p *ProjectTree) watchErrs() {
	for err := range p.watcher.Errors {
		log.Printf("Watcher received error %s", err)
	}
}

// watch waits for events from p.watcher.  For each event, the tree will
// spin off a goroutine to update the its children.
//
// Events are processed in separate goroutines to help us keep up with
// rapidly occurring events, e.g. in the case of a `git checkout` that
// touches many, many files and directories.  It doesn't completely
// prevent UI lock up, but it mitigates it some.
func (p *ProjectTree) watch() {
	for e := range p.watcher.Events {
		switch e.Op {
		case fsnotify.Write, fsnotify.Create, fsnotify.Remove, fsnotify.Rename:
			go p.update(e.Name)
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
// So far, though, I haven't seen a situation where fsnotify overwhelms
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

func (p *ProjectTree) SetProject(project settings.Project) {
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
