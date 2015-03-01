package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/jpillora/watcher/watcher"
)

const help = `
	Usage: watcher [--dir DIR] [--delay DELAY] program ...args

	Watches for changes to all files in DIR (defaults to the current
	directory). After each change, program will be restarted.
	Restarts are debounced by DELAY (defaults to '100ms').

	Read more:
	https://github.com/jpillora/watcher

`

func main() {
	//flag stuff
	dir := flag.String("dir", "./", "Working directory (defaults to current)")
	delay := flag.Duration("delay", 100*time.Millisecond, "Duration to delay each restart")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, help)
	}
	flag.Parse()
	args := flag.Args()
	//start!
	w, err := watcher.New(*dir, *delay, args)
	if err != nil {
		log.Fatal(err)
	}
	//show info prints
	w.Info = true
	//stop on CTRL+C
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)
	go func() {
		<-sig
		w.Stop()
	}()

	//stop and block
	if err := w.Start(); err != nil {
		log.Fatal(err)
	}
}
