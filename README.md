![TMSU](http://tmsu.org/images/tmsu.png)

[![Build Status](https://travis-ci.org/oniony/TMSU.svg?branch=master)](https://travis-ci.org/oniony/TMSU)
[![Go Report Card](https://goreportcard.com/badge/github.com/oniony/TMSU)](https://goreportcard.com/report/github.com/oniony/TMSU)

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

Packages
--------

Thanks to the efforts of contributors using these platforms, packages are available
for the following GNU/Linux distributions:

  * Ubuntu
    - Stable <https://launchpad.net/~tmsu/+archive/ubuntu/ppa>
    - Daily <https://launchpad.net/~tmsu/+archive/ubuntu/daily>
  * Arch
    - Stable <https://aur.archlinux.org/packages/tmsu/>

These packages are not maintained by me and I cannot guarantee their content.

Binary
------

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

TMSU is written in Go: <http://www.golang.org/>

Much of the functionality the program provides is made possible by the FUSE and
Sqlite3 libraries, their Go bindings and the Go language standard library.

  * Website: <http://tmsu.org/>
  * Project: <https://github.com/oniony/TMSU/>
  * Wiki: <https://github.com/oniony/TMSU/wiki>
  * Issue tracker: <https://github.com/oniony/TMSU/issues>
  * Mailing list: <http://groups.google.com/group/tmsu>

Release Notes
=============

v0.7.1
------

  * VFS now uses relative paths -- thanks to [foxcpp](https://github.com/foxcpp)
  * Support for Blake2b-256 fingerprints -- thanks to [foxcpp](https://github.com/foxcpp)
  * Hidden files are no longer tagged by default when tagging recursively. To
    include hidden files use the `--include-hidden` option -- thanks to
    [foxcpp](https://github.com/foxcpp)

v0.7.0
------

  *Note: this release changes how symbolic links are handled. See below.*

  * TMSU now compiles for Mac O/S. (Thanks to https://github.com/pguth.)
  * The VFS no longer lists files alongside the tag directories under `tags`.
    Instead there are `results` directories at each level, within which you can
    find the set of symbolic links to the tagged files.
  * Symbolic links are now followed by default. This means that if you tag a
    symbolic link, the target file is tagged instead. To instruct TMSU to not
    follow symbolic links (previous behaviour) use the new `--no-dereference`
    option on the relevant subcommands.
  * Added new setting `symlinkFingerprintAlgorithm` to allow the fingerprint
    algorithm for symbolic links to be configured separately to regular files.
  * By default duplicate files will now be reported when tagging. A new setting
    `reportDuplicates` can be used to turn this off.
  * Slashes are now permitted within tags and values, useful for recording URLs.
    In the virtual filesystem, similar looking Unicode characters are used in
    their place.
  * Added `--where` option to `tag` subcommand to allow tags to be applied to
    the set of files matching a query.
  * The VFS tags directory will now relist tags that have values so that
    multiple values can be specified, e.g. tags/color/=red/color/=blue.
  * It is now possible to list tags that use a particular value with the new
    `--value` option on the 'tags' subcommand.
  * Made it possible to upgrade the database schema between releases.
  * Added `--count` option to `untagged`.
  * Bug fixes.

v0.6.1
------

  * Fixed crash when opening an empty tag directory in the VFS.

v0.6.0
------

  *Note: this release changes the database schema by adding additional columns
  to the 'implication' table. TMSU will automatically upgrade your database
  upon first use but you may wish to take a backup beforehand.*

  * Relaxed restrictions on tag and value names, allowing punctuation and
    whitespace characters. Problematic characters can be escaped with backslash.
  * Values are no longer automatically deleted when no longer used: it is now
    up to you to manage them.
  * Added --force option to 'tag' subcommand to allow tagging of missing or
    permission denied paths and broken symlinks.
  * 'imply' now creates tags if necessary (and 'autoCreateTags' is set).
  * Performance improvements to the virtual filesystem.
  * Fixed 'too many SQL variables' when merging tags applied to lots of files.
  * Added --name option to 'tags' to force printing of name even if there is
    only a single file argument, which is useful when using xargs.
  * Replaced 'stats' subcommand with 'info' subcommand (with --stats and --usage
    options for tag statics and usage counts respectively).
  * Included a set of scripts for performing filesystem operations whilst
    keeping the TMSU database up to date. If you wish to use these scripts
    I recommend you alias them to simpler names, e.g. 'trm'.
    - `tmsu-fs-rm`     Removes files from the filesystem and TMSU
    - `tmsu-fs-mv`     Moves a file in the filesystem and updates TMSU
    - `tmsu-fs-merge`  Merges files (deleting all but the last)
  * Tag values can now be renamed, deleted and merged using the new --value
    option on the corresponding subcommands.
  * Tag values can now be used in implications.
  * Tag values can be explicitly created: tmsu tag --create =2015. (It may
    be necessary to enclose the argument in quotes depending upon your shell.)
  * It is no longer possible to add a circular tag implication. (These were
    not correctly applied anyway. An alias facility will be provided in a later
    version.)
  * The output of 'files' can now be sorted in various ways using --sort.
  * Case insensitive queries can now be performed with the --ignore-case option
    on the 'files' subcommand.
  * Added integration tests covering CLI.
  * Bug fixes.

v0.5.2
------

  * Fixed bug where concurrent access to the virtual filesystem would cause
    a runtime panic.

v0.5.1
------

  * Fixed bug with database initialization when .tmsu directory does not
    already exist.

v0.5.0
------

  *Note: This release has some important changes, including the renaming of
  some options, the introduction of local databases and a switch from absolute
  to relative paths in the database. Please read the following release notes
  carefully.*

  * The --untagged option on the 'files' and 'status' subcommands has been
    replaced by a new 'untagged' subcommand, which should be more intuitive.
  * The --all option on the 'files', 'tags' and 'values' subcommands has been
    removed. These commands now list the full set of files/tags/values when run
    without arguments. For the 'tags' subcommand this replaces the previous
    behaviour of listing tags for the files in the working directory: use 'tmsu
    tags *' for approximately the previous behaviour.
  * The 'repair' subcommand --pretend short option has changed from -p to -P (so
    that -p can be recycled for --path).
  * The 'repair' subcommand's argument now specify paths to search for moved
    files and no longer limit how much of the database is repaired. A new --path
    argument is provided for reducing the repair to a portion of the database.
  * A new --manual option on the 'repair' subcommand allows targetted repair of
    moved files or directories.
  * The exclamation mark character (!) is no longer permitted within a tag or
    value name. Please rename tags using the 'rename' command. (Value names will
    need to be updated manually using the Sqlite3 tooling.)
  * Added --colour option to the 'tags' subcommand to highlight implied tags.
  * 'tag' subcommand will, by default, no longer explicitly apply tags that are
    already implied (unless the new --explicit option is specified).
  * Added subcommand aliases, e.g. 'query' for 'files'.
  * It is now possible to tag a broken symbolic link: instead of an error this
    will now be reported as a warning.
  * It is now possible to remove tags with values via the VFS.
  * 'tag' subcommand can tag multiple files with different tags by reading from
    standard input by passing an argument of '-'.
  * TMSU will now automatically use a local database in .tmsu/db in working
    directory or any parent. The new 'init' subcommand allows a new local
    database to be initialized. See [Switching Databases](https://github.com/oniony/TMSU/wiki/Switching%20Databases).
  * Paths are now stored relative to the .tmsu directory's parent rather than as
    absolute paths. This allows a branch of the filesystem to be moved around,
    shared or archived whilst preserving the tagging information. Existing
    absolute paths can be converted by running a manual repair:

        tmsu repair --manual / /

  * Added 'config' subcommand to view and amend settings.
  * The 'help' subcommand now wraps textual output to fit the terminal.
  * Rudimentary Microsoft Windows support (no virtual filesystem yet).
  * TMSU can now be built without the Makefile.
  * Bug fixes.

v0.4.3
------

  * Fixed unit-test problems.

v0.4.2
------

  * Fixed bug where 'dynamic:MD5' and 'dynamic:SHA1' fingerprint algorithms
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
  * 'tags' and 'values' subcommands now tabulate output, by default, when run
    from terminal.
  * Added ability to configure which fingerprint algorithm to use.
  * Implied tags now calculated on-the-fly when the database is queried. This
    results in a (potentially) smaller database and ability to have updates to
    the implied tags affect previously tagged files.
  * Added --explicit option to 'files' and 'tags' subcommands to show only
    explicit tags (omitting any implied tags).
  * Added --path option to 'files' subcommand to retrieve just those files
    matching or under the path specified.
  * Added --untagged option to 'files' subcommand which, when combined with
    --path, will also include untagged files from the filesystem at the
    specified path.
  * Removed the --recursive option from the 'files' subcommand which was flawed:
    use 'tmsu files query | xargs find' instead.
  * Added ability to configure whether new tags and values are automatically
    created or not or a per-database basis.
  * Added --unmodified option to 'repair' subcommand to force the recalculation
    of fingerprints of unmodified files.
  * Renamed --force option of 'repair' subcommand to --remove.
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
  * 'tag' subcommand now allows tags to be created up front.
  * 'copy' and 'imply' subcommands now support multiple destination tags.
  * Improved 'stats' subcommand.
  * Added man page.
  * Added script to allow the virtual filesystem to be mounted via the
    system 'mount' command or on startup via the fstab.
  * Bug fixes.

- - -

Copyright 2011–2017 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
