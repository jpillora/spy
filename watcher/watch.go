package watcher

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jpillora/ansi"
	"gopkg.in/fsnotify.v1"
)

//Watcher takes a directory and a program. It runs
//this program whenever files in directory change.
type Watcher struct {
	//enable Info or Debug stdout logging
	Info, Debug bool
	//include hidden directories and files
	IncludeHidden bool

	dir      string
	watched  map[string]bool
	proc     *process
	watcher  *fsnotify.Watcher
	watching chan bool
	log      *log.Logger
}

//NewWatcher creates a new Watcher
func New(dir string, delay time.Duration, args []string) (*Watcher, error) {
	w := &Watcher{}

	w.dir = dir
	w.watched = make(map[string]bool)
	w.watching = make(chan bool)

	w.log = log.New(bluify(0), "watcher ", log.Ldate|log.Ltime|log.Lmicroseconds)

	var err error
	w.proc, err = newProcess(w, args, delay)
	if err != nil {
		return nil, err
	}
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (w *Watcher) Start() error {
	defer w.watcher.Close()

	//change to dir
	if err := os.Chdir(w.dir); err != nil {
		return err
	}
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	//then watch root!
	w.watch(dir)
	w.info("Watching '%s'", dir)

	//queue watcher to close
	go w.handleEvents()

	//start the process [manager]
	go w.proc.start()
	defer w.proc.stop()

	//block
	<-w.watching
	return nil
}

func (w *Watcher) Stop() {
	close(w.watching)
}

func (w *Watcher) watch(path string) {
	if w.IncludeHidden && isHidden(path) {
		return
	}
	w.debug("watch '%s'", path)
	if _, ok := w.watched[path]; ok {
		return
	}
	w.watched[path] = true
	w.watcher.Add(path)
	//recurse
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		if f.IsDir() {
			w.watch(path + "/" + f.Name())
		}
	}
}

func (w *Watcher) unwatch(path string) bool {
	if _, ok := w.watched[path]; !ok {
		return false
	}
	//w.watcher.Remove seems to be implicitly called
	delete(w.watched, path)
	if path == w.dir {
		close(w.watching)
	}
	return true
}

func (w *Watcher) handleEvents() {
	for {
		select {
		case event := <-w.watcher.Events:
			w.handleEvent(event)
		case err := <-w.watcher.Errors:
			w.debug("watch error %s", err)
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {

	if !w.IncludeHidden && isHidden(event.Name) {
		return
	}
	//cant stat - doesn't exist anymore
	if event.Op&fsnotify.Remove == fsnotify.Remove {
		if !w.unwatch(event.Name) {
			//remove file? restart
			w.proc.restart()
		}
		return
	}
	//only CREATE or WRITE are viewed as change events
	if event.Op&fsnotify.Create != fsnotify.Create &&
		event.Op&fsnotify.Write != fsnotify.Write {
		return
	}

	s, err := os.Stat(event.Name)
	if err != nil {
		w.debug("file stat error: %s", err)
		return
	}

	if s.IsDir() {
		w.watch(event.Name)
	} else {
		w.debug("file changed: %s", event.Name)
		w.proc.restart()
	}
}

func (w *Watcher) info(f string, args ...interface{}) {
	if w.Info {
		w.log.Printf(f, args...)
	}
}

func (w *Watcher) debug(f string, args ...interface{}) {
	if w.Debug {
		w.log.Printf(f, args...)
	}
}

//helpers

type bluify int

func (g bluify) Write(p []byte) (n int, err error) {
	os.Stdout.Write(ansi.Set(ansi.Blue))
	os.Stdout.Write(p)
	os.Stdout.Write(ansi.Set(ansi.Reset))
	return len(p), nil
}

func isHidden(path string) bool {
	return strings.HasPrefix(filepath.Base(path), ".")
}
