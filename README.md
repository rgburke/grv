# GRV - Git Repository Viewer [![Build Status](https://travis-ci.org/rgburke/grv.svg?branch=master)](https://travis-ci.org/rgburke/grv)

GRV is a terminal based interface for viewing git repositories. It allows
refs, commits and diffs to be viewed, searched and filtered. The behaviour
and style can be customised through configuration. A query language can
be used to filter refs and commits, see the [Documentation](#documentation)
section for more information.

GRV is currently under development and not feature complete.

## Demo

![Demo](doc/grv.gif)

## Documentation

Documentation for GRV is available [here](doc/documentation.md)

## Build instructions

GRV depends on the following libraries:

 - libncursesw
 - libreadline
 - libcurl
 - cmake (to build libgit2)

Building GRV on OSX requires homebrew and for readline to be installed using homebrew.

To install GRV run:

```
go get -d github.com/rgburke/grv/cmd/grv
cd $GOPATH/src/github.com/rgburke/grv
make install
```

This will install GRV to `$GOPATH/bin`. A static libgit2 will be built and
included in GRV when built this way. Alternatively if libgit2 0.25 is
installed on your system GRV can be built normally:

```
go install ./cmd/grv
```
