![TMSU](http://tmsu.org/images/tmsu.png)

![Build](https://github.com/oniony/TMSU/actions/workflows/build.yml/badge.svg)

NOTE: TMSU has moved to ![Codeberg](https://codeberg.org/oniony/TMSU).

The ![Github repository](https://github.com/oniony/TMSU) is now a mirror of the
Codeberg repository.

# Overview

TMSU is a tool for tagging your files. It provides a simple command-line utility
for applying and managing tags, and a virtual filesystem to allow you to access
your files by tag, or combination of tags, from any application, even those with
a graphical user interface.

TMSU does not alter your files in any way: you simply gain an additional view,
which you can mount where you like, based upon the tags you set up. If you decide
to stop using TMSU, your files will be exactly as they were before you started.

# Usage

## Initialise Database

Before you can start tagging, you will need to initialise a TMSU database in the
folder in which you wish to use tags:

    $ cd example
    $ tmsu init

This will create a database under `.tmsu` in that directory that will be used
automatically whenever you are under that directory (much like a version control
system such as Git).

## Tagging

One can tag a file by specifying the file and the list of tags to apply:

    $ tmsu tag banana.jpg fruit art year=2015

This will tag the file, `banana.jpg` as `fruit`, `art` and `year`. Additionally,
the `year` tag has a value `2015` that adds more detail.

Alternatively, one can instead specify the tags first:

    $ tmsu tag --tags="fruit still-life art" banana.jpg apple.png

This is particularly useful when piping a list of files:

    $ find . -name "*.mp3" | xargs tmsu tag --tags "music"

Will tag all of the files with an `.mp3` extension `music`.

## Querying

One can query for files with or without particular tags using `and`, `or` and
`not`:

    $ tmsu files fruit or still-life
    $ tmsu files fruit and art
    $ tmsu files not still-life and art
    $ tmsu files trippy and (art or music)

The `and` keyword is actually implied, so it can be omitted:

    $ tmsu files fruit still-life
    $ tmsu files fruit art
    $ tmsu files not still-life art
    $ tmsu files trippy (art or music)

For tags that include values, it is also possible to query based upon the values
of these tags:

    $ tmsu files author=somebody
    $ tmsu files "year < 2000"

The `<`, `>`, `=`, `!=` operators can also be expressed as `lt`, `gt`, `eq` and `ne`.

## Virtual File System

TMSU lets a tagged database be mounted as a virtual file system, allowing the tagged
files to be accessed by tag rather than directory from any application.

Mount the virtual filesystem to an empty directory:

    $ mkdir mp
    $ tmsu mount mp

Query the files by tag:

    $ ls "mp/queries/fruit and art"

This will automatically create a `fruit and art` directory in the virtual file system
containing links to all of the files in the database matching those tags.

This can be accomplished in a graphical application too by navigating to the `queries`
directory in the file chooser and typing in the query: the virtual file system will
automatically detect this and create the directory containing links to the matching
files.

## More Information

The integrated help can be used to get usage information:

    $ tmsu help
    $ tmsu help tags

# Installing

## Package

TODO

## Binary

Binary builds for a limited number of architectures and operating system
combinations are available at <https://codeberg.org/oniony/TMSU/releases>.

You will need to ensure that both FUSE and Sqlite3 are installed for the
program to function.

1. Install the binary

    Copy the program binary. The location may be different for your operating
    system:

        $ sudo cp bin/tmsu /usr/bin

2. Optional: Zsh completion

    Copy the Zsh completion file to the Zsh site-functions directory:

        $ cp misc/zsh/_tmsu /usr/share/zsh/site-functions

## Source

If you would like to build from the source code then please see `COMPILING.md`.

# About

TMSU itself is written and maintained by [Paul Ruane](mailto:Paul Ruane <paul.ruane@oniony.com>).

# Release Notes

See <https://codeberg.org/oniony/TMSU/releases>.

- - -

Copyright 2011-2025 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
