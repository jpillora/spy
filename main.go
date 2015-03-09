package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jpillora/spy/spy"
)

const help = `
	Usage: spy [options] program ...args

	program (along with it's args) is initially
	run and then it is restarted with every file
	change. program will always be run from the
	current working directory.

	Options:

	--inc INCLUDE, Describes a path to files to
	watch. Use ** to describe any number of
	directories. Use * to describe any file name.
	For example, you could watch all Go source
	files with "**/*.go" or all	JavaScript source
	files in './lib/' with "lib/**/*.js".

	--exc EXCLUDE, Describes a path to files not
	to watch. Inverse of INCLUDE.

	--dir DIR, Watches for changes to all files in
	DIR (defaults to the current directory). After
	each change, program will be restarted.

	--delay DELAY, Restarts are debounced by DELAY
	(defaults to '0.5s').

	-color -c, Color of spy log text. Can choose
	between: c,m,y,k,r,g,b,w (defaults to
	"g" green)

	-v, Enable verbose logging

	Read more:
	https://github.com/jpillora/spy

`

func main() {
	//flag stuff
	dir := flag.String("dir", "./", "")
	inc := flag.String("inc", "", "")
	exc := flag.String("exc", "", "")
	color := flag.String("color", "", "")
	c := flag.String("c", "", "")
	verbose := flag.Bool("v", false, "")
	delay := flag.Duration("delay", 500*time.Millisecond, "")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, help)
	}
	flag.Parse()
	args := flag.Args()

	//flag pkg lacks alias support
	if *c != "" {
		*color = *c
	}

	//start!
	w, err := spy.New(*dir, *color, *delay, args)
	if err != nil {
		fmt.Printf("\n\t%s\n", err)
		flag.Usage()
		os.Exit(1)
	}
	//show info prints
	w.Info = true
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
	//start and block
	if err := w.Start(); err != nil {
		log.Fatal(err)
	}
}
