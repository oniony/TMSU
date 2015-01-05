Compiling TMSU From Source
==========================

These instructions are for compiling TMSU from the source code. If you are using
a downloaded binary then they can be safely ignored.

1. Installing Go

    TMSU is written in the Go programming language. To compile from source you must
    first install Go and the packages that TMSU depends upon. You can get that from
    the Go website:

    * <http://www.golang.org/>

    Go can be installed per the instructions on the Go website or it may be
    available in the package management system that comes with your operating
    system.

2. Set up the Go path

    Go (as of verison 1.1) requires the GOPATH environment variable be set for the
    'go get' command to function. You will need to set up a path for Go packages to
    live if you do not already have one:

        $ mkdir $HOME/gopath
        $ export GOPATH=$HOME/gopath

3. Install the dependent packages.

    These will be installed to your $GOPATH directory.

        $ go get -u github.com/mattn/go-sqlite3
        $ go get -u github.com/hanwen/go-fuse/fuse

4. Clone the TMSU respository:

    To clone the current stable release branch:

        $ git clone -b v0.4 https://github.com/oniony/TMSU.git

    Active development takes place on the 'master' branch. This branch is
    subject to build failures and breaking changes but will have the latest
    functionality and improvements:

        $ git clone https://github.com/oniony/TMSU.git

5. Make the project

        $ cd tmsu
        $ make

    This will compile to 'bin/tmsu' within the working directory.

6. Install the project

        $ sudo make install

    This will install TMSU to '/usr/bin/tmsu'.

    It will also install the Zsh completion to '/usr/share/zsh/site-functions'
    and mount wrapper to '/usr/sbin'.

    To change the paths used override the variables at the top of the Makefile.

- - -

Copyright 2011-2015 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
