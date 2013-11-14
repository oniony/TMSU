Overview
========

TMSU is a program that allows you to organise your files by associating them
with tags. It provides a tool for managing these tags and a virtual filesystem
to allow tag-based access to your files.

TMSU's virtual filesystem does not store your files: it merely provides an
alternative, tag-based view of your files stored elsewhere in the filesystem.
That way you have the freedom to choose the most suitable filesystem for
storage whilst still benefiting from tag-based access.

Usage
=====

A command overview and details on how to use each command are available via the
integrated help:

    $ tmsu help

Full documentation is maintained online on the wiki:

  * <http://bitbucket.org/oniony/tmsu/wiki>

Installing
==========

Binary builds for a limited number of architectures and operating system
combinations are available:

  * <http://bitbucket.org/oniony/tmsu/downloads>

You will need to ensure that both FUSE and Sqlite3 are installed for the
program to function. These packages are typically available with your
operating system's package management system.

1. Install the binary

Copy the program binary. The location may be different for your operating
system:

    $ sudo cp bin/tmsu /usr/bin

2. Optional: Zsh completion

Copy the Zsh completion file to the Zsh site-functions directory:

    $ cp misc/zsh/_tmsu /usr/share/zsh/site-functions

Compiling
=========

The following steps are for compiling from source.

1. Installing Go

    TMSU is written in the Go programming language. To compile from source you must
    first install Go and the packages that TMSU depends upon. You can get that from
    the Go website:

    * <http://www.golang.org/>

    Go can be installed per the instructions on the Go website or it may be
    available in the package management system that comes with your operating
    system.

2. Set Up the Go Path

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

        $ hg clone https://bitbucket.org/oniony/tmsu

5. Make the project

        $ cd tmsu
        $ make

    This will compile to 'bin/tmsu' within the working directory.

6. Install the project

        $ sudo make install

    This will install TMSU to '/usr/bin/tmsu'.

    It will also install the Zsh completion to '/usr/share/zsh/site-functions'.

    To change the paths used override the environment variables in the Makefile.

About
=====

TMSU itself is written and maintained by Paul Ruane <paul@tmsu.org>.

TMSU is written in Go: <http://www.golang.org/>

Much of the functionality the program provides is made possible by the FUSE and
Sqlite3 libraries, their Go bindings and the Go language standard library.

  * Website: <http://tmsu.org/>
  * Project: <http://bitbucket.org/oniony/tmsu>
  * Wiki: <http://bitbucket.org/oniony/tmsu/wiki>
  * Mailing list: <http://groups.google.com/group/tmsu>

Release Notes
=============

v0.3.0 (in development)
-----------------------

  Note: This release changes what tag names are allowed. To ensure the tag
  names in your existing databases are still valid, please run the following
  script:

      $ cp ~/.tmsu/default.db ~/.tmsu/default.db.bak
      $ sqlite3 -init misc/db-upgrades/clean_tag_names.sql ~/.tmsu/default.db

  * Added support for file queries, e.g. 'fish and chips and (mushy-peas or
    ketchup)'
  * Added support for file queries in the virtual filesystem: create a query
    directory under the 'queries' directory.
  * Added global option --database for specifying database location.
  * Added ability to rename and delete tags via the virtual filesystem.
  * 'tag' command now allows tags to be created up front.
  * 'copy' and 'imply' commands now support multiple destination tags.
  * Improved 'stats' command.
  * Added man page.
  * Added script to allow the virtual filesystem to be mounted via the
    system mount command or on startup via the fstab.
  * Bug fixes.

v0.2.2
------

  * Fixed virtual filesystem.

v0.2.1
------

  * Fixed bug where excluding multiple tags would return incorrect results.
  * Fixed Go 1.1 compilation problems. 

v0.2.0
------

  * Added support for tag implications, e.g. tag 'a' implies 'b'. New 'imply'
    command for managing these.
  * Added --force option to 'repair' command to remove missing files (and
    associated taggings) from the database.
  * Added --from option to 'tag' command to allow tags to copied from one file
    to another. e.g. 'tmsu tag -f a b' will apply file b's tags to file a.
    ('tag -r -f a a' will recursively retag a directory's contents.)
  * Added --directory option to 'status' command to stop it recursively
    processing directory contents.
  * Added --print0 option to 'files' command to allow use with xargs.
  * Added --count option to 'tags' and 'files' command to list tag/file count
    rather than names.
  * Bug fixes and unit-test improvements.

- - -

Copyright 2011 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
