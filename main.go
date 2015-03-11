package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/jpillora/spy/spy"
)

var VERSION string = "0.0.0-src" //set via ldflags

var help = `
	Usage: spy [options] program ...args

	program (along with its args) is initially run and then restarted with every file
	change. program will always be run from the current working directory.

	Options:

	--inc INCLUDE - Describes a path to files to watch. Use ** to wildcard directories
	and use * to wildcard file names. For example, you could watch all Go source files
	with "--inc **/*.go" or all	JavaScript source files in ./lib/ with
	"--inc lib/**/*.js".

	--exc EXCLUDE - Describes a path to files not to watch. Inverse of INCLUDE. For
	example, you could exclude your static front-end directory with "--exc static/".

	--dir DIR, Watches for changes to all files in DIR (defaults to the current
	directory). After each change, program will be restarted.

	--delay DELAY, Restarts are throttled by DELAY (defaults to '0.5s'). For example,
	a "save all open files" action could trigger multiple file changes, though only
	a single restart since these changes would all fall inside the DELAY period.

	--color -c, Color of log text. Can choose between: c,m,y,k,r,g,b,w.

	--verbose -v, Enable verbose logging

	--quiet -q, Disable all logging

	--version, Display version (` + VERSION + `)

	Read more:
	https://github.com/jpillora/spy
`

func main() {
	//flag stuff
	dir := flag.String("dir", "./", "")
	inc := flag.String("inc", "", "")
	exc := flag.String("exc", "", "")
	version := flag.Bool("version", false, "")
	color := flag.String("color", "", "")
	c := flag.String("c", "", "")
	verbose := flag.Bool("verbose", false, "")
	v := flag.Bool("v", false, "")
	quiet := flag.Bool("quiet", false, "")
	q := flag.Bool("q", false, "")
	delay := flag.Duration("delay", 500*time.Millisecond, "")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, help)
	}
	flag.Parse()
	args := flag.Args()

	if *version {
		fmt.Println(VERSION)
		os.Exit(1)
	}

	//flag pkg lacks alias support
	if *v {
		*verbose = true
	}
	if *q {
		*quiet = true
	}
	if *c != "" {
		*color = *c
	}

	//start!
	w, err := spy.New(*dir, *color, *delay, args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\n\t%s\n", err)
		flag.Usage()
		os.Exit(1)
	}
	//show info prints
	w.Info = !*quiet
	w.Debug = *verbose
	w.Include = *inc
	w.Exclude = *exc
	//stop on CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	go func() {
		<-sig
		w.Stop()
	}()

	//start watching
	w.Start()
	//block
	if err := w.Wait(); err != nil {
		os.Exit(1)
	}
}
