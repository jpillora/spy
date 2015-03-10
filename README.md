# Spy

Spy is a simple, general purpose, cross platform, file spy written in Go (Golang). Spy takes a `directory` and a `program` and runs `program` whenever a file in `directory` changes.

### Install

Binaries

See [Releases](https://github.com/jpillora/spy/releases)

<!--
spy_1.0.0_windows_amd64.zip
spy_1.0.0_windows_386.zip
spy_1.0.0_linux_arm.tar.gz
spy_1.0.0_linux_amd64.tar.gz
spy_1.0.0_linux_386.tar.gz
spy_1.0.0_i386.deb
spy_1.0.0_darwin_amd64.zip
spy_1.0.0_darwin_386.zip
spy_1.0.0_armhf.deb
spy_1.0.0_amd64.deb
-->

Source

```
go get -v github.com/jpillora/spy
```

*Currently, `spy` does not fully support windows, as it uses process groups to ensure all sub-processes have exited between restarts. A pull request which implements the `process_win.go` file would be appreciated.*

### Usage

```
$ spy --help

	Usage: spy [options] program ...args

	program (along with it's args) is initially
	run and then it is restarted with every file
	change. program will always be run from the
	current working directory.

	Options:

	--inc INCLUDE - Describes a path to files to
	watch. Use ** to wildcard directories and use
	* to wildcard file names. For example, you could 
	watch all Go source files with "--inc **/*.go"
	or all	JavaScript source files in ./lib/
	with "--inc lib/**/*.js".

	--exc EXCLUDE - Describes a path to files not
	to watch. Inverse of INCLUDE. For example, you
	could exclude your static front-end directory
	with "--exc static/".

	--dir DIR, Watches for changes to all files in
	DIR (defaults to the current directory). After
	each change, program will be restarted.

	--delay DELAY, Restarts are debounced by DELAY
	(defaults to '0.5s').

	-color -c, Color of spy log text. Can choose
	between: c,m,y,k,r,g,b,w (defaults to
	"g" green)

	--verbose -v, Enable verbose logging

	--version, Display version

	Read more:
	https://github.com/jpillora/spy

```

### Examples

Go - Auto restart server

``` sh
$ cd $GOPATH/github.com/user/repo
$ spy go run main.go
spy 2015/03/02 01:40:40.248290 Watching .
2015/03/02 01:40:40 listening on 8080...
# ... change a file ...
spy 2015/03/02 01:40:43.996842 Restarting...
2015/03/02 01:40:44 listening on 8080...
```

Go - Auto re-run tests

```
$ spy go test
```

Node

```
$ spy node server.js
```

Bash

```
$ spy bash program.sh
```

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
