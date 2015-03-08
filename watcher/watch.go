package watcher

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jpillora/ansi"
	"gopkg.in/fsnotify.v1"
)

//Watcher takes a directory and a program. It runs
//this program whenever files in directory change.
type Watcher struct {
	//enable Info or Debug stdout logging
	Info, Debug bool
	//inclusion or exclusion filters
	Include, Exclude string
	//include hidden directories
	IncludeHidden bool

	dir      string
	dirs     map[string]bool
	proc     *process
	watcher  *fsnotify.Watcher
	watching chan bool
	log      *log.Logger
	matcher  *matcher
}

//NewWatcher creates a new Watcher
func New(dir string, delay time.Duration, args []string) (*Watcher, error) {
	w := &Watcher{}

	w.dir = dir
	w.dirs = make(map[string]bool)
	w.watching = make(chan bool)

	w.log = log.New(bluify, "watcher ", log.Ldate|log.Ltime|log.Lmicroseconds)

	var err error
	w.proc, err = newProcess(w, args, delay)
	if err != nil {
		return nil, err
	}
	w.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	w.matcher = &matcher{include: true}
	return w, nil
}

func (w *Watcher) Start() error {
	defer w.watcher.Close()

	dir, err := filepath.Abs(w.dir)
	if err != nil {
		return err
	}
	w.dir = dir

	//initialize matchers
	if w.Include != "" {
		w.matcher.glob(dir + "/" + w.Include)
	} else if w.Exclude != "" {
		w.matcher.glob(dir + "/" + w.Exclude)
		w.matcher.include = false
	}

	//watch root path!
	w.watch(dir)
	w.info("Watching %s", shorten(dir))

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
	if !w.matcher.matchDir(path) {
		return
	}
	if err := w.watcher.Add(path); err != nil {
		w.debug("watch  failed: %s (%s)", path, err)
		return
	}
	w.debug("watch: %s", path)
	w.dirs[path] = true
	//recurse
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		if f.IsDir() {
			w.watch(path + "/" + f.Name())
		}
	}
}

func (w *Watcher) handleEvents() {
	for {
		select {
		case event := <-w.watcher.Events:
			go w.handleEvent(event)
		case err := <-w.watcher.Errors:
			w.debug("watch error %s", err)
		}
	}
}

func (w *Watcher) handleEvent(event fsnotify.Event) {
	// w.debug("event: %s", event)
	path := event.Name
	if !w.matcher.matchFile(path) {
		return
	}
	//cant stat - doesn't exist anymore
	if event.Op&fsnotify.Remove == fsnotify.Remove ||
		event.Op&fsnotify.Rename == fsnotify.Rename {
		if _, ok := w.dirs[path]; ok {
			//root dir removed!
			if path == w.dir {
				close(w.watching)
			}
		} else {
			//matched file deleted
			w.debug("file deleted: %s", path)
			w.proc.restart()
		}
		return
	}
	//only CREATE or WRITE are viewed as change events
	if event.Op&fsnotify.Create != fsnotify.Create &&
		event.Op&fsnotify.Write != fsnotify.Write {
		return
	}

	s, err := os.Stat(path)
	if err != nil {
		w.debug("file stat error: %s", err)
		return
	}

	if s.IsDir() {
		w.watch(path)
	} else {
		w.debug("file changed: %s", path)
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

type blueWriter int

func (g blueWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(ansi.Set(ansi.Blue))
	os.Stdout.Write(p)
	os.Stdout.Write(ansi.Set(ansi.Reset))
	return len(p), nil
}

var bluify blueWriter

func shorten(path string) string {
	wd, err := os.Getwd()
	if err != nil {
		return path
	}
	rpath, err := filepath.Rel(wd, path)
	if err != nil {
		return path
	}
	return rpath
}
