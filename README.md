# Spy

Spy is a simple, general purpose, cross platform, file system watcher written in Go (Golang). Spy takes a `directory` and a `program` and runs `program` whenever a file in `directory` changes.

### Install

**Binaries**

See [Releases](https://github.com/jpillora/spy/releases)

**Source**

```
go get -v github.com/jpillora/spy
```

### Quick start

Auto-restart your Go server

``` sh
$ spy go run main.go
spy 2015/03/02 01:40:40.248290 Watching .
2015/03/02 01:40:40 listening on 8080...
# ... change a file ...
spy 2015/03/02 01:40:43.996842 Restarting...
2015/03/02 01:40:44 listening on 8080...
```

### Usage

```
$ spy --help

	Usage: spy [options] program ...args

	program (along with its args) is initially run and then restarted with every file
	change. program will always be run from the current working directory.

	Options:

	--dir DIR, Watches for changes to all files in DIR (defaults to the current
	directory). After each change, program will be restarted.

	--inc INCLUDE - Describes a path to files to watch. Use ** to wildcard directories
	and use * to wildcard file names. This path will be made relative to DIR. For example,
	you could watch all Go source files with "--inc **/*.go" or all	JavaScript source
	files in ./lib/ with "--inc lib/**/*.js".

	--exc EXCLUDE - Describes a path to files not to watch. Inverse of INCLUDE. For
	example, you could exclude your static front-end directory with "--exc static".

	--delay DELAY, Restarts are throttled by DELAY (defaults to '0.5s'). For example,
	a "save all open files" action might trigger multiple file changes, though only
	a single restart would occur since these changes would all fall inside the DELAY
	period.

  --match MATCH - Describes a pattern for path to files to watch. For example
	"--match '(go|txt|po)$'".

	--color -c, Color of log text. Can choose between: c,m,y,k,r,g,b,w.

	--verbose -v, Enable verbose logging

	--quiet -q, Disable all logging

	--version, Display version (` + VERSION + `)

	Read more:
	https://github.com/jpillora/spy

```

### More examples

Auto-rerun tests *with green spy logs*

```
$ spy -c green go test
```

Auto-restart Node server *only when `.js` files change*

```
$ spy --inc "**/*.js" node server.js
```

Auto-rerun a shell script

```
$ spy ./program.sh
```

Set regexp for matching

```
$ spy --match '(go|html|js|css)$' --exc node_modules
```

### Issues

Q: "too many files open"

A: http://stackoverflow.com/a/34645/977939

### Todo

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
