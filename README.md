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

A command overview and details on how to use each command are available via the
integrated help:

    $ tmsu help

Documentation is maintained online on the wiki:

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

    To clone the current stable release branch:

        $ hg clone -r release https://bitbucket.org/oniony/tmsu

    Active development takes place on the default branch. This branch is subject
    to build failures and breaking changes but will have the latest
    functionality and improvements:

        $ hg clone https://bitbucket.org/oniony/tmsu

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

About
=====

TMSU itself is written and maintained by Paul Ruane <paul@tmsu.org>.

The creation of TMSU is motivation in itself, but if you should feel inclinded
to make a small gift via Bitcoin then it shall be gratefully received:

  * 1TMSU5TL3Yj6AGP7Wq6uahTfkTSX2nWvM

TMSU is written in Go: <http://www.golang.org/>

Much of the functionality the program provides is made possible by the FUSE and
Sqlite3 libraries, their Go bindings and the Go language standard library.

  * Website: <http://tmsu.org/>
  * Project: <http://bitbucket.org/oniony/tmsu>
  * Wiki: <http://bitbucket.org/oniony/tmsu/wiki>
  * Mailing list: <http://groups.google.com/group/tmsu>

Release Notes
=============

v0.5.0 (in development)
------

  *Note: This release prohibts the inclusion of the exclamation mark (!)
  character within tag and value names. Please use 'tmsu rename' to change the
  names of any existing tags that include this character.*

  * Added --untagged option to 'status' command for showing untagged files only.
  * Added --colour option to the 'tags' command to highlight implied tags.
  * 'tag' command will, by default, no longer explicitly apply tags that are
    already implied (unless the new --explicit option is specified).
  * Rudimentary Microsoft Windows support (no virtual filesystem yet).
  * Disallowed use of '!' within tag and value names.
  * It is now possible to tag a broken symbolic link: instead of an error this
    will now be reported as a warning.

v0.4.2
------

  * Fixed bug where 'dynamic:MD5' and 'dynamic:SHA1' fingerprint algorrithms
    were actually using SHA256.

v0.4.1
------

  * Tag values are now shown as directories in the virtual filesystem.

v0.4.0
------

  *Note: This release changes the database schema to facilitate tag values. To
  upgrade your existing v0.3.0 database please run the following:*

    $ cp ~/.tmsu/default.db ~/.tmsu/default.db.backup
    $ sqlite3 -init misc/db-upgrade/0.3_to_0.4.0.sql ~/.tmsu/default.db .q

  * Added support for tag values, e.g. 'tmsu tag song.mp3 country=uk' and the
    querying of files based upon these values, e.g. 'year > 2000'.
  * 'tags' and 'values' commands now tabulate output, by default, when run
    from terminal.
  * Added ability to configure which fingerprint algorithm to use.
  * Implied tags now calculated on-the-fly when the database is queried. This
    results in a (potentially) smaller database and ability to have updates to the
    implied tags affect previously tagged files.
  * Added --explicit option to 'files' and 'tags' commands to show only
    explicit tags (omitting any implied tags).
  * Added --path option to 'files' command to retrieve just those files matching
    or under the path specified.
  * Added --untagged option to 'files' command which, when combined with --path,
    will also include untagged files from the filesystem at the specified path.
  * Removed the --recursive option from the 'files' command which was flawed:
    use 'tmsu files query | xargs find' instead.
  * Added ability to configure whether new tags and values are automatically
    created or not or a per-database basis.
  * Added --unmodified option to 'repair' command to force the recalculation
    of fingerprints of unmodified files.
  * Renamed --force option of 'repair' command to --remove.
  * Added support for textual comparison operators: 'eq', 'ne', 'lt', 'gt',
    'le' and 'ge', which do not need escaping unlike '<', '>', &c.
  * Improved Zsh completion with respect to tag values.
  * Significant performance improvements.
  * Removed support for '-' operator: use 'not' instead.
  * Bug fixes.

v0.3.0
------

  *Note: This release changes what tag names are allowed. To ensure the tag
  names in your existing databases are still valid, please run the following
  script:*

    $ cp ~/.tmsu/default.db ~/.tmsu/default.db.backup
    $ sqlite3 -init misc/db-upgrade/clean_tag_names.sql ~/.tmsu/default.db

  * Added support for file queries, e.g. 'fish and chips and (mushy-peas or
    ketchup)'.
  * Added support for file query directories in the virtual filesystem.
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

Copyright 2011-2014 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
