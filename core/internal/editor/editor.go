// This is free and unencumbered software released into the public
// domain.  For more information, see <http://unlicense.org> or the
// accompanying UNLICENSE file.

package editor

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/nelsam/vidar/fsw"
	"github.com/nelsam/vidar/input"
	"github.com/nelsam/vidar/theme"
	"github.com/nelsam/vidar/ui"
)

type Runner interface {
	Enqueue(op func())
}

type Creator interface {
	Editor() (ed ui.TextEditor, err error)
}

type Watcher interface {
	Watch(path string) (fsw.EventHandler, error)
}

type CodeEditor struct {
	box ui.TextEditor

	runner ui.Runner
	events fsw.EventHandler

	mu       sync.RWMutex
	filepath string

	syntaxTheme theme.Theme
	layers      []input.SyntaxLayer

	lastModified time.Time
}

// NewCodeEditor creates a CodeEditor using the passed in values.
func NewCodeEditor(creator Creator, syntaxTheme theme.Theme, file string, w Watcher) (*CodeEditor, error) {
	ed, err := creator.Editor()
	if err != nil {
		return nil, err
	}
	e := &CodeEditor{
		box:         ed,
		filepath:    file,
		syntaxTheme: syntaxTheme,
		runner:      creator.Runner(),
	}
	go e.init(w)
	return e, nil
}

func (e *CodeEditor) init(w Watcher) {
	eh, err := w.Watch(e.filepath)
	if os.IsNotExist(err) {
		eh, err = e.waitForFileCreate(w)
	}
	if err != nil {
		// TODO: report this failure in the UI
		log.Printf("Could not watch %s for changes: %s", e.filepath, err)
	}
	e.events = eh
	e.load()
	e.watch(w)
}

func (e *CodeEditor) waitForFileCreate(w Watcher) (eh fsw.EventHandler, err error) {
	dir := filepath.Dir(e.filepath)
	if err := os.MkdirAll(dir, 0750|os.ModeDir); err != nil {
		return err
	}
	eh, err = w.Watch(dir)
	if err != nil {
		return err
	}
	defer func() {
		if err == nil {
			if err = eh.Add(e.filepath); err != nil {
				return
			}
			err = eh.Remove(dir)
		}
	}()

	for {
		ev, err := eh.Next()
		if err != nil {
			return nil, err
		}
		if ev.Path == e.filepath && ev.Op&fsw.Create == fsw.Create {
			return eh, nil
		}
	}
}

func (e *CodeEditor) watch(w Watcher) {
	if e.events == nil {
		// TODO: report this to the UI
		log.Printf("CodeEditor.watch called without an event handler set")
		return
	}
	defer e.watcher.Remove(e.filepath)
	for {
		ev, err := e.events.Next()
		if err != nil {
			log.Printf("Error from watcher: %s", err)
			e.load()
			return
		}
		switch ev.Op {
		case fsw.Write:
			e.load()
		}
	}
}

func (e *CodeEditor) Renamed(newPath string) {
	e.events.Remove(e.filepath)
	e.events.Add(newPath)
	e.load()
}

func (e *CodeEditor) load() {
	f, err := os.Open(e.filepath)
	if os.IsNotExist(err) {
		return
	}
	if err != nil {
		// TODO: report this failure to the UI
		log.Printf("Error opening file %s: %s", e.filepath, err)
		return
	}
	defer f.Close()
	finfo, err := f.Stat()
	if err != nil {
		// TODO: report this failure to the UI
		log.Printf("Error statting file %s: %s", e.filepath, err)
		return
	}
	b, err := ioutil.ReadAll(f)
	if err != nil {
		// TODO: report this failure to the UI
		log.Printf("Error reading file %s: %s", e.filepath, err)
		return
	}
	e.runner.Enqueue(func() {
		// TODO: handle differences between the FS and the editor - don't overwrite!
		if string(e.box.Text()) == string(newText) {
			return
		}
		e.box.SetText([]rune(newText))
	})
}

func (e *CodeEditor) Text() []rune {
	return e.box.Text()
}

func (e *CodeEditor) Apply(ed) {

}

func (e *CodeEditor) MTime() time.Time {
	finfo, err := os.Stat(e.Filepath())
	if err != nil {
		return time.Now()
	}
	return finfo.ModTime()
}

func (e *CodeEditor) Filepath() string {
	return e.filepath
}

func (e *CodeEditor) SetSyntaxLayers(layers []input.SyntaxLayer) {
	defer e.syntaxTheme.Rainbow.Reset()
	sort.Slice(layers, func(i, j int) bool {
		return layers[i].Construct < layers[j].Construct
	})
	e.layers = layers
	var spans []ui.Span
	for _, l := range layers {
		highlight, found := e.syntaxTheme.Highlights[l.Construct]
		if !found {
			highlight = e.syntaxTheme.Rainbow.Next()
		}
		for _, s := range l.Spans {
			spans = append(spans, ui.Span{
				Highlight: highlight,
				Start:     uint64(s.Start),
				Length:    uint64(s.End - s.Start),
			})
		}
	}
	e.box.SetSpans(spans...)
}

func (e *CodeEditor) SyntaxLayers() []input.SyntaxLayer {
	return e.layers
}
