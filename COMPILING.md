Compiling TMSU From Source
==========================

These instructions are for compiling TMSU from the source code. If you are using
a downloaded binary then they can be safely ignored.

1. Installing Go

    * Note: macOS users can skip down to the macOS section as it includes
      installing Go.

    TMSU is written in the Go programming language. To compile from source you must
    first install Go:

    * <http://www.golang.org/>

    Go can be installed per the instructions on the Go website or it may be
    available in the package management system that comes with your operating
    system.

3. Set up the Go path

    Go, as of verison 1.1, requires the `GOPATH` environment variable be set for
    the 'go get' command to function. You will need to set up a path for Go
    packages to live if you do not already have one. Please see the following
    page for details on how to set this up:

    * <http://golang.org/cmd/go/#hdr-GOPATH_environment_variable>

4. Clone the TMSU respository:

    To clone the current stable release branch:

        git clone -b v0.6 https://github.com/oniony/TMSU.git

    Active development takes place on the 'master' branch. This branch is
    subject to build failures and breaking changes but will have the latest
    functionality and improvements:

        git clone https://github.com/oniony/TMSU.git

Now follow the next steps according to your operating system.

Linux
-----

4. Build and install

    make
    sudo make install

    or, to compile manually:

    mkdir bin
    cd src/github.com/oniony/TMSU
    go build -o ../../../../bin/tmsu .

    This will build the binary and copy it to `/usr/bin`, as well as installing
    Zsh completion, a `mount` wrapper and the manual page. To adjust the paths
    please edit the `Makefile`.

macOS
-----

1. Install Homebrew

    Homebrew allows you to install additional software and command-line tools
    that macOS doesn't ship with. This is used to get GNU versions of cp and
    find, and use them instead of macOS's BSD equivalents.

    * <https://brew.sh/>

    Homebrew will prompt for your password and install the Command Line Tools
    for XCode if required. Conveniently, this also includes GNU `make` which
    we'll also need to compile.

    Homebrew will finish with instructions for putting the Homebrew downloaded
    binaries into your local shell's PATH. Follow those instructions.
    Note: `mount.tmsu`, a warpper for `tmsu mount`, installs into
    `/usr/local/sbin`; you may need to adjust your `$PATH` if you'd like to use
    it.

2.  Install Go:

    brew install go

    Follow the instructions under Linux, for using `git clone` to pull in TMSU.
    The source code will be pulled into the TMSU directory.

3. Install GNU cp and find

    These are utilities found in Homebrew's `coreutils` and `findutils`
    packages:

    brew install coreutils findutils

4. Build and install

    cd TMSU/
    make
    make install

    This will build the binary and copy it to `/usr/local/bin`, as well as
    installing Zsh completion, and a manual page. To adjust the paths please
    edit the `Makefile`.

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

Copyright 2011-2018 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
