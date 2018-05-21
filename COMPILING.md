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

3. Clone the TMSU respository:

    To clone the current stable release branch:

        git clone -b v0.6 https://github.com/oniony/TMSU.git

    Active development takes place on the 'master' branch. This branch is
    subject to build failures and breaking changes but will have the latest
    functionality and improvements:

        git clone https://github.com/oniony/TMSU.git

Now follow the next steps according to your operating system.

Linux
-----

4. Install the dependent packages

    These will be installed to your GOPATH directory (see previous step).

        go get -u golang.org/x/crypto/blake2b
        go get -u github.com/mattn/go-sqlite3
        go get -u github.com/hanwen/go-fuse/fuse

5. Build and install

        make
        sudo make install

    This will build the binary and copy it to `/usr/bin`, aswell as installing
    Zsh completion, a `mount` wrapper and the manual page. To adjust the paths
    please edit the `Makefile`.

Windows
-------

4. Set up MinGW

    To compile the dependent package, go-sqlite3, it is necessary to set up a MinGW
    environment.

    There are various options available, which you can peruse these at the download
    page, but I found Msys2 pretty painless to install:

    * <http://mingw-w64.org/>

5. Install GCC

    TMSU uses the go-sqlite3 library, which in turn requires Sqlite3 to be compiled.
    To do this you need to install GCC from within the MinGW environment. If you
    installed Msys2, above, then this can be accomplised using the following pacman
    command from an Msys2 terminal:

        pacman -S mingw64/mingw-w64-x86_64-gcc

    If you used another option, you will need to install this package either through
    the graphical installer or package manager provided, if it is not installed by
    default.

6. Install the dependent packages

    This will be installed to your GOPATH directory (see step 2). The following
    command has to be run from the MinGW terminal (e.g. Msys2) otherwise it will fail
    to compile Sqlite3:

        go get -u github.com/mattn/go-sqlite3
        go get -u golang.org/x/crypto/blake2b


7. Set the path

    Within the MinGW terminal, and from the directory where you cloned TMSU, configure
    the GOPATH environment variable:

        export GOPATH=$PWD:$GOPATH
    
8. Build and install

    Now run the following command:

        go build -o tmsu.exe github.com/oniony/TMSU

    This will build `tmsu.exe` to the working directory.

- - -

Copyright 2011-2017 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
