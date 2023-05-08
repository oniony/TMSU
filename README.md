![TMSU](http://tmsu.org/images/tmsu.png)

# Overview

TMSU is a tool for tagging your files. It provides a simple command-line utility
for applying tags and a virtual filesystem to give you a tag-based view of your
files from any application.

TMSU does not alter your files in any way: you simply gain an additional view,
which you can mount where you like, based upon the tags you set up.

# Usage

## Initialise Database

Before you can get tagging, you'll need to initialise a TMSU database in a
folder in which you wish to use tags:

    $ cd example
    $ tmsu init

This will create a database under `.tmsu` that will be used automatically whenever
you are under that directory.

## Tagging

You can tag a file by specifying the file and the list of tags to apply:

    $ tmsu tag banana.jpg fruit art year=2015

This will tag your file, `banana.jpg` as `fruit`, `art` and `year`.
Additionally, the `year` tag has a value `2015` that adds more detail. (We'll come
back to tag values later.)

You can instead specify the tags first:

    $ tmsu tag --tags="fruit still-life art" banana.jpg apple.png

This is particularly useful when piping a list of files:

    $ find . -name "*.mp3" | xargs tmsu tag --tags "music"

## Querying

You can query for files with or without particular tags:

    $ tmsu files fruit or still-life
    $ tmsu files fruit and art
    $ tmsu files not still-life and art
    $ tmsu files trippy and (art or music)

The `and` keyword is implied, so you can just leave it out:

    $ tmsu files fruit art
    $ tmsu files not still-life art
    $ tmsu files trippy (art or music)

Where you've used values on your tags, you can query based upon the values of the tags:

    $ tmsu files author=somebody
    $ tmsu files "year < 2000"

## Virtual File System

The virtual file system lets you mount your tagged files as a file system, allowing you to
access your files by tags from any other application.

Mount the virtual filesystem to an empty directory:

    $ mkdir mp
    $ tmsu mount mp
    
## More Information

A subcommand overview and detail on how to use each subcommand is available via the
integrated help:

    $ tmsu help
    $ tmsu help tags

Documentation is maintained online on the wiki:

  * <https://github.com/oniony/TMSU/wiki>

# Installing

## Packages

Thanks to the efforts of contributors using these platforms, packages are available
for the following GNU/Linux distributions:

  * Ubuntu
    - Stable <https://launchpad.net/~tmsu/+archive/ubuntu/ppa>
    - Daily <https://launchpad.net/~tmsu/+archive/ubuntu/daily>
  * Arch
    - Stable <https://aur.archlinux.org/packages/tmsu/>
  * Nix/NixOS
    - Stable <https://search.nixos.org/packages?query=tmsu&show=tmsu>
    - Unstable <https://search.nixos.org/packages?query=tmsu&show=tmsu&channel=unstable>

These packages are not maintained by me and I cannot guarantee their content.

## Binary

Binary builds for a limited number of architectures and operating system
combinations are available:

  * <https://github.com/oniony/TMSU/releases>

You will need to ensure that both FUSE and Sqlite3 are installed for the
program to function. These packages are typically available with your
operating system's package management system. (If you install TMSU using one
of the above packages, these should be installed automatically.)

1. Install the binary

    Copy the program binary. The location may be different for your operating
    system:

        $ sudo cp bin/tmsu /usr/bin

2. Optional: Zsh completion

    Copy the Zsh completion file to the Zsh site-functions directory:

        $ cp misc/zsh/_tmsu /usr/share/zsh/site-functions

## From Source

If you would rather build from the source code then please see `COMPILING.md`
in the root of the repository.

# About

TMSU itself is written and maintained by [Paul Ruane](mailto:Paul Ruane <paul@tmsu.org>).

# Release Notes

The rewrite in Rust is currently underway. The result of this will be v1.0.0.

Previous version history is available on the old `master` branch.

- - -

Copyright 2011-2023 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
