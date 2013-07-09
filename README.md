Overview
========

TMSU is an application that allows you to organise your files by associating
them with tags. It provides a tool for managing these tags and a virtual file-
system to allow tag based access to your files.

TMSU's virtual file system does not store your files: it merely provides an
alternative, tagged based view of your files stored elsewhere in the file-
system. That way you have the freedom to choose the most suitable file-system
for storage whilst still benefiting from tag based access.

Usage
=====

A command overview and details on how to use each command are available via the
integrated help:

    $ tmsu help

Full documentation is maintained online on the wiki:

  * <http://bitbucket.org/oniony/tmsu/wiki>

Downloading
===========

Binary builds for a limited number of architectures and operating system
combinations are available:

  * <http://bitbucket.org/oniony/tmsu/downloads>

You will need to ensure that both FUSE and Sqlite3 are installed for the
program to function.

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

        $ go get github.com/mattn/go-sqlite3
        $ go get github.com/hanwen/go-fuse/fuse

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

  * Website: <http://www.tmsu.org/>
  * Project: <http://bitbucket.org/oniony/tmsu>
  * Wiki: <http://bitbucket.org/oniony/tmsu/wiki>
  * Mailing list: <http://groups.google.com/group/tmsu>

TMSU is written in Go: <http://www.golang.org/>

TMSU itself is written and maintained by Paul Ruane <paul@tmsu.org>, however
much of the functionality it provides is made possible by the FUSE and Sqlite3
libraries, their Go bindings and, of course, the Go language standard library.

Release Notes
=============

v0.3.0
------

  * Feature: tag values.

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

v0.1.2
------

  * Removed use of Sqlite bulk insert functionality as the version of Sqlite
    in mainstream Linux distributions does not have this functionality.

v0.1.1
------

  * Fixed panic when using the 'rename' command with the incorrect number of
    arguments.
  * 'status' command now handles case where file is replaced with directory
    or vice-versa.
  * Fixed bug in 'status' command where mixing of relative and absolute files
    was missing some moved files.

v0.1.0
------

This version aims to improve the performance of the program by removing auto-
matic tag inheritence which has proven slow on larger databases and slow file-
systems (e.g. network file-systems). Instead there is now a choice of how to
handle directory contents:

  * Add the nested files to the database using the --recursive option on the
    'tag' and 'untag' commands.
  * Dynamically discover directory contents using the --recursive option on
    'files' command.

Adding directory contents using the --recursive option on the 'tag' command
will be slower and will result in a larger database. However querying of the
database will be faster as TMSU will not need to examine the file-system. This
option also means that duplicate files will be identified.

Dynamically discovering directory conents using the --recursive option on the
'files' command will be slower as TMSU will need to scan the file-system and
duplicate files will not be identified as these files are not added to the
database. However tagging files will be faster and the resultant database will
be smaller.

IMPORTANT: Please back up your database then upgrade it using the upgrade
script. The 'repair' step may take a while to run as every file is reexamined
to populate the new columns.

IMPORTANT: If you have been following 'tip' there is a separate upgrade
script called `tip_to_0.1.0.sql`. Due to a bug in the previous version
please re-run this even if you have previously upgraded otherwise you may end
up with duplicate entries in the `file_tag` table. Please read through the
script before running it.

    $ cp ~/.tmsu/default.db ~/.tmsu/default.db.bak  # back up
    $ sqlite3 -init misc/db-upgrades/0.0.9_to_0.1.0.sql ~/.tmsu/default.db
    $ tmsu repair

  * 'files' command no longer finds files that inherit a tag from the file-
    system on-the-fly unless run with the new --recursive (-r) option. New
    --directory (-d) and --file (-f) options to limit output to just files or
    directories. In addition the new --top (-t) and --leaf (-l) options show the
    top-most matching entries (excluding items from matching directories) or
    bottom-most matching entries (excluding parent directories) respectevly.
  * 'tag' and 'untag' now sport a --recursive (-r) option for tagging or
    untagging directory contents.
  * Improved command-line parsing: now supports global options, short options
    and mixed option ordering. Options with arguments can now be specified any-
    where on the command line.
  * 'repair' command rewritten to fix bugs.
  * Fingerprints for directories no longer calculated. (Recursively tag
    files instead to detect file duplicates.)
  * Removed the 'export' command. (Sqlite tooling has better facilities.)
  * Tags containing '/' are no longer legal.
  * Fixed bug that prevented file lists from being shown sorted alphanumerically.
  * The 'tag' command no longer identifies modified files. (Use 'repair'
    instead.)
  * The 'mount' command now has a '--allow-other' option which allows other
    users to access the mounted file-system.
  * Updated Zsh completion.
  * Improved error messages.
  * Improved unit-test coverage.
  * Minor bug fixes.

v0.0.9
------

  * Fixed bug which caused process hosting the virtual file-system to crash if
    a non-existant tag directory is 'stat'ed.
  * Untagged files now inherit parent directory tags.

v0.0.8
------

Files can now be tagged within tagged directories. Files within tagged
directories will inherit the directory's tags.

  * Fixed bug with 'untag' command when non-existant tag is specified.
  * Updated with respect to go-fuse API change. 
  * 'mount' command now lists mount points if invoked without arguments.
  * Improved 'mount' command help.
  * 'rename' command now validates destination tag name.
  * Fixed bug with 'unmount --all' returning an error if there are no mounts.
  * Removed dependency upon 'pgrep'; now accesses proc-fs directotly for mount
    information.
  * 'files' command will now show files that inherit the specified tags
    ('--explicit' option turns this off).
  * 'tags' command will now shows inherited tags ('--explicit' turns this off).
  * 'status' command now reports inherited tags.
  * 'stats' command formatting updated.
  * Other minor bug fixes.

v0.0.7
------

  Files larger than 5MB now use a different fingerprinting algorithm where the
  fingerprint is produced by taking a 500KB of the start, middle and end of the
  file. This should dramatically improve performance, especially on slow file-
  systems.

  NOTE: it is advisable to run 'repair' after upgrading to this version to update
        large files with fingerprints produced using the new algorithm.

v0.0.6
------

IMPORTANT: This version adds a column to one of the database tables to record
the file's modification timestamep. The new code will not work with an
existing TMSU database until it has been upgraded.

To upgrade the database, run the upgrade script using the Sqlite3 tooling:

    $ sqlite3 -init misc/db-upgrade/0.0.5_to_0.0.6.sql

It is also advisable to run 'repair' after upgrading which will populate the
new column and also fix the directory fingerprints which have a new algorithm.

  * Upgraded to Go 1.
  * Added 'repair' command to fix up database when files are moved or modified.
  * Added 'version' command.
  * Added 'copy' command (duplicates a tag).
  * Added modification timestamp which is used in preference to fingerprint.
  * Command output now includes 'tmsu' to make it clear where output is coming
    from when piping.
  * Relative paths now calculated more accurately.
  * Zsh completion now supports tags containing colons.
  * 'status' command performance and functionality improvements.
  * Added directory fingerprinting.
  * Added 'repair' command to Zsh completion.
  * The 'files' command now allows tags to be excluded by prefixing them with a
    minus, e.g. -jazz.

- - -

Copyright 2011 Paul Ruane

Copying and distribution of this file, with or without modification,
are permitted in any medium without royalty provided the copyright
notice and this notice are preserved.  This file is offered as-is,
without any warranty.
