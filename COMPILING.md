Compiling TMSU From Source
==========================

These instructions are for compiling TMSU from the source code. If you are using
a downloaded binary then they can be safely ignored.

1. Installing Go

    TMSU is written in the Go programming language. To compile from source you must
    first install Go:

    * <http://www.golang.org/>

    Go can be installed per the instructions on the Go website or it may be
    available in the package management system that comes with your operating
    system.

2. Set up the Go path

    Go, as of verison 1.1, requires the `GOPATH` environment variable be set for
    the 'go get' command to function. You will need to set up a path for Go
    packages to live if you do not already have one. Please see the following
    page for details on how to set this up:

    * <http://golang.org/cmd/go/#hdr-GOPATH_environment_variable>

3. Install the dependent packages.

    These will be installed to your GOPATH directory (see previous step).

    Unix like operating systems:

        go get -u github.com/mattn/go-sqlite3
        go get -u github.com/hanwen/go-fuse/fuse

    Microsoft Windows:

        go get -u github.com/mattn/go-sqlite3

4. Clone the TMSU respository:

    To clone the current stable release branch:

        git clone -b v0.4 https://github.com/oniony/TMSU.git

    Active development takes place on the 'master' branch. This branch is
    subject to build failures and breaking changes but will have the latest
    functionality and improvements:

        git clone https://github.com/oniony/TMSU.git

5. Build and install

    Unix like operating systems:

        make
        sudo make install

    This will build the binary and copy it to `/usr/bin`, aswell as installing
    Zsh completion, a `mount` wrapper and the manual page. To adjust the paths
    please edit the `Makefile`.

    Microsoft Windows:

        go build tmsu

    This will build `tmsu.exe` to the working directory.

- - -

Copyright 2011-2015 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
