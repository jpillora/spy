# Watcher

Watcher is cross platform, general purpose, file watcher written in Go (Golang). Watcher takes a `directory` and a `program` and runs `program` whenever a file in `directory` changes.

### Install

```
go get -v github.com/jpillora/watcher
```

:warning: `watcher` uses process groups to ensure all sub-processes have exited between restarts, so only Unix systems currently work. PR to add Windows accepted!

### Usage

```
$ watcher --help

	Usage: watcher [--dir DIR] [--delay DELAY] program ...args

	Watches for changes to all files in DIR (defaults to the current
	directory). After each change, program will be restarted.
	Restarts are debounced by DELAY (defaults to '100ms').

	Read more:
	https://github.com/jpillora/watcher
```

### Example

Go

```
$ cd $GOPATH/github.com/user/repo
$ watcher go run main.go
watcher 2015/03/02 01:40:40.248290 Watching '.../github.com/user/repo'
2015/03/02 01:40:40 listening on 8080...
watcher 2015/03/02 01:40:43.996842 Restarting...
2015/03/02 01:40:44 listening on 8080...
```

Node

```
$ watcher node server.js
```

### Todo

* Include/Exclude glob options
* Extract Unix specific code into a conditional build file
* Port Unix code to Windows

#### MIT License

Copyright © 2015 &lt;dev@jpillora.com&gt;

Permission is hereby granted, free of charge, to any person obtaining
a copy of this software and associated documentation files (the
'Software'), to deal in the Software without restriction, including
without limitation the rights to use, copy, modify, merge, publish,
distribute, sublicense, and/or sell copies of the Software, and to
permit persons to whom the Software is furnished to do so, subject to
the following conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED 'AS IS', WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT.
IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY
CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT,
TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE
SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
