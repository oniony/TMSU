![TMSU](http://tmsu.org/images/tmsu.png)

Important
=========

This branch is an experiment in reimplementing TMSU in the Rust programming language.

  1. I have not committed to switch to Rust yet, this is an evaluation so this
     experiment may be discontinued at any time without notice.
  2. This is very much a work-in-progress so please do not use this implementation
     as your daily driver.

Overview
========

TMSU is a tool for tagging your files. It provides a simple command-line utility
for applying tags and a virtual filesystem to give you a tag-based view of your
files from any other program.

TMSU does not alter your files in any way: they remain unchanged on disk, or on
the network, wherever your put them. TMSU maintains its own database and you
simply gain an additional view, which you can mount where you like, based upon
the tags you set up.

Usage
=====

Before you can get tagging, you'll need to initialise a TMSU database:

    $ cd ~
    $ tmsu init

This database will be used automatically whenever you are under that
directory. In this case we created one under the home directory.

You can tag a file by specifying the file and the list of tags to apply:

    $ tmsu tag banana.jpg fruit art year=2015

Or you can apply tags to multiple files:

    $ tmsu tag --tags="fruit still-life art" banana.jpg apple.png

You can query for files with or without particular tags:

    $ tmsu files fruit and not still-life

Mount the virtual filesystem to an empty directory:

    $ mkdir mp
    $ tmsu mount mp
    
A subcommand overview and detail on how to use each subcommand is available via the
integrated help:

    $ tmsu help
    $ tmsu help tags

Documentation is maintained online on the wiki:

  * <https://github.com/oniony/TMSU/wiki>

Installing
==========

TODO

From Source
-----------

If you would rather build from the source code then please see `COMPILING.md`
in the root of the repository.

About
=====

TMSU itself is written and maintained by [Paul Ruane](mailto:Paul Ruane <paul@tmsu.org>).

The creation of TMSU is motivation in itself, but if you should feel inclinded
to make a small gift via Pledgie or Bitcoin then it shall be gratefully received:

  * <https://pledgie.com/campaigns/31085>
  * `1TMSU5TL3Yj6AGP7Wq6uahTfkTSX2nWvM`

TMSU is written in Rust: <http://www.rust-lang.org/>

  * Website: <http://tmsu.org/>
  * Project: <https://github.com/oniony/TMSU/>
  * Wiki: <https://github.com/oniony/TMSU/wiki>
  * Issue tracker: <https://github.com/oniony/TMSU/issues>
  * Mailing list: <http://groups.google.com/group/tmsu>

Release Notes
=============

v1.0.0 (in development)
------

TODO

- - -

Copyright 2011-2017 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
