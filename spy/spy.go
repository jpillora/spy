package spy

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/jpillora/ansi"
	"gopkg.in/fsnotify.v1"
)

//Spy takes a program and a directory. It runs
//program whenever files in directory change.
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
	watching chan error
	closer   sync.Once
	log      *log.Logger
	matcher  *matcher
}

//NewWatcher creates a new Spy
func New(dir string, logColor string, delay time.Duration, args []string) (*Spy, error) {
	s := &Spy{}

	dir, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}

	s.dir = dir
	s.dirs = make(map[string]bool)
	s.watching = make(chan error, 1)

	var logWriter io.Writer
	if logColor != "" {
		logWriter = newColorWriter(logColor)
	} else {
		logWriter = os.Stdout
	}

	s.log = log.New(logWriter, "spy ", log.Ldate|log.Ltime|log.Lmicroseconds)

	s.proc, err = newProcess(s, args, delay)
	if err != nil {
		return nil, err
	}
	s.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	s.matcher = &matcher{include: true}
	return s, nil
}

func (s *Spy) Start() {
	//initialize matchers
	if s.Include != "" {
		s.matcher.glob(join(s.dir, s.Include))
	} else if s.Exclude != "" {
		s.matcher.glob(join(s.dir, s.Exclude))
		s.matcher.include = false
	} /* else match all! */

	//watch root path!
	s.watch(s.dir)
	s.info("Watching %s", shorten(s.dir))
	//queue spy to close
	go s.handleEvents()
	//start the process [manager]
	go s.proc.start()
}

func (s *Spy) Run() error {
	s.Start()
	return s.Wait()
}

func (s *Spy) Wait() error {
	err := <-s.watching
	if err != nil {
		s.info("Error: %v", err)
	}
	return err
}

func (s *Spy) stopWith(err error) {
	s.closer.Do(func() {
		s.proc.stop()
		s.watcher.Close()
		s.watching <- err
		close(s.watching)
	})
}

func (s *Spy) Stop() {
	s.stopWith(nil)
}

func (s *Spy) watch(path string) {
	if !s.matcher.matchDir(path) {
		return
	}
	if err := s.watcher.Add(path); err != nil {
		s.debug("watch failed: %s", err)
		s.stopWith(fmt.Errorf("%s (%s)", err, path))
		return
	}
	s.dirs[path] = true
	s.debug("watch #%d: %s", len(s.dirs), path)
	//recurse
	files, _ := ioutil.ReadDir(path)
	for _, f := range files {
		if f.IsDir() {
			s.watch(join(path, f.Name()))
		}
	}
}

//runs in a goroutine
func (s *Spy) handleEvents() {
	for {
		select {
		case event := <-s.watcher.Events:
			go s.handleEvent(event)
		case err := <-s.watcher.Errors:
			if err != nil {
				s.debug("watch error %s", err)
			}
		}
	}
}

func (s *Spy) handleEvent(event fsnotify.Event) {
	// s.debug("event: %s", event)
	path := event.Name
	if !s.matcher.matchFile(path) {
		return
	}
	//cant stat - doesn't exist anymore
	if event.Op&fsnotify.Remove == fsnotify.Remove ||
		event.Op&fsnotify.Rename == fsnotify.Rename {
		if _, ok := s.dirs[path]; ok {
			//root dir removed!
			if path == s.dir {
				s.stopWith(fmt.Errorf("spy directory removed (%s)", path))
			}
		} else {
			//matched file deleted
			s.debug("file deleted: %s", path)
			s.proc.restart()
		}
		return
	}
	//only CREATE or WRITE are viewed as change events
	if event.Op&fsnotify.Create != fsnotify.Create &&
		event.Op&fsnotify.Write != fsnotify.Write {
		return
	}

	info, err := os.Stat(path)
	if err != nil {
		s.debug("file stat error: %s", err)
		return
	}

	if info.IsDir() {
		s.watch(path)
	} else {
		s.debug("file changed: %s", path)
		s.proc.restart()
	}
}

func (s *Spy) info(f string, args ...interface{}) {
	if s.Info {
		s.log.Printf(f, args...)
	}
}

func (s *Spy) debug(f string, args ...interface{}) {
	if s.Debug {
		s.log.Printf(f, args...)
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

//path.join, though keep the trailing slash
func join(paths ...string) string {
	s := path.Join(paths...)
	if strings.HasSuffix(paths[len(paths)-1], "/") {
		s += "/"
	}
	return s
}
