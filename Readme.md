[![Build Status](https://travis-ci.org/ericaro/compgen.png?branch=master)](https://travis-ci.org/ericaro/compgen) [![GoDoc](https://godoc.org/github.com/ericaro/compgen?status.svg)](https://godoc.org/github.com/ericaro/compgen)

This library is fully `go gettable`.

# Compgen

`compgen` is a Go package to write bash completion to command. like

    $ git comm<TAB>
    $ git commit


Even though it can be any command, it makes more sense if both program share a bit of code. It is even possible to integrate both programs into a single one: the completion program and the execution program are just the same.

If you use any CLI framework there are chances that you can "introspect" the CLI struct (options, sub commands etc.) and use it to build the completion program.

To register a program that is self completing you just need to:

    complete -C cmd cmd 

Or copy the above statement in a file into  `/etc/bash_completion.d/`

# License

help is available under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0.html).

# Branches

master: [![Build Status](https://travis-ci.org/ericaro/compgen.png?branch=master)](https://travis-ci.org/ericaro/compgen) against go versions:

  - 1.2
  - 1.3
  - tip

dev: [![Build Status](https://travis-ci.org/ericaro/compgen.png?branch=dev)](https://travis-ci.org/ericaro/compgen) against go versions:

  - 1.2
  - 1.3
  - tip


