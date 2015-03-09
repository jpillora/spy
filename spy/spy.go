package spy

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/jpillora/ansi"
	"gopkg.in/fsnotify.v1"
)

//Spy takes a directory and a program. It runs
//this program whenever files in directory change.
type Spy struct {
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

//NewWatcher creates a new Spy
func New(dir string, color string, delay time.Duration, args []string) (*Spy, error) {
	w := &Spy{}

	w.dir = dir
	w.dirs = make(map[string]bool)
	w.watching = make(chan bool)

	w.log = log.New(newColorWriter(color), "spy ", log.Ldate|log.Ltime|log.Lmicroseconds)

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

func (w *Spy) Start() error {
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

	//queue spy to close
	go w.handleEvents()

	//start the process [manager]
	go w.proc.start()
	defer w.proc.stop()

	//block
	<-w.watching
	return nil
}

func (w *Spy) Stop() {
	close(w.watching)
}

func (w *Spy) watch(path string) {
	if !w.matcher.matchDir(path) {
		return
	}
	if err := w.watcher.Add(path); err != nil {
		w.info("watch failed: %s (%s)", path, err)
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

func (w *Spy) handleEvents() {
	for {
		select {
		case event := <-w.watcher.Events:
			go w.handleEvent(event)
		case err := <-w.watcher.Errors:
			w.debug("watch error %s", err)
		}
	}
}

func (w *Spy) handleEvent(event fsnotify.Event) {
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

func (w *Spy) info(f string, args ...interface{}) {
	if w.Info {
		w.log.Printf(f, args...)
	}
}

func (w *Spy) debug(f string, args ...interface{}) {
	if w.Debug {
		w.log.Printf(f, args...)
	}
}

//helpers

type colorWriter ansi.Attribute

func newColorWriter(letter string) colorWriter {
	switch letter {
	case "c":
		return colorWriter(ansi.Cyan)
	case "m":
		return colorWriter(ansi.Magenta)
	case "y":
		return colorWriter(ansi.Yellow)
	case "k":
		return colorWriter(ansi.Black)
	case "r":
		return colorWriter(ansi.Red)
	case "g":
		return colorWriter(ansi.Green)
	case "b":
		return colorWriter(ansi.Blue)
	case "w":
		return colorWriter(ansi.White)
	default:
		return colorWriter(ansi.Green)
	}
}

func (c colorWriter) Write(p []byte) (n int, err error) {
	os.Stdout.Write(ansi.Set(ansi.Attribute(c)))
	os.Stdout.Write(p)
	os.Stdout.Write(ansi.Set(ansi.Reset))
	return len(p), nil
}

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
