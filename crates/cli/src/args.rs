// Copyright 2011-2025 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

use crate::rendering::Separator;
use clap::{ArgAction, Parser, Subcommand};
use std::path::PathBuf;

#[derive(Parser)]
#[command(about = "TMSU", version, long_about = None)]
#[command(disable_colored_help = true)]
pub struct Args {
    #[clap(short = 'D', long = "database")]
    pub database: Option<PathBuf>,

    #[clap(short = 'v', long = "verbose", action = ArgAction::Count, default_value_t = 0)]
    pub verbosity: u8,

    #[arg(
        short = '0',
        long = "print0",
        help = "delimit files with a NUL character rather than newline",
        default_value_t = false
    )]
    print0: bool,

    #[command(subcommand)]
    pub command: Commands,
}

impl Args {
    pub fn parse() -> Self {
        Parser::parse()
    }

    pub fn separator(&self) -> Separator {
        match self.print0 {
            true => Separator::Nul,
            false => Separator::Newline,
        }
    }
}

#[derive(Subcommand, Debug)]
pub enum Commands {
    #[command(
        about = "Show database information",
        long_about = "Show database paths and metrics."
    )]
    Info,

    #[command(
        about = "Initialize a new database",
        long_about = "Initializes new, empty databases.

Typically this command is run without arguments to create a new, local database at .tmsu/db in the
current working directory. Such a database will be used automatically when TMSU is used at or
below this directory.

If PATH is specified, then a database will be created at this specific path instead. Such a database
will only be subsequently used when TMSU is run with an explicit database path via the --database
global option or the TMSU_DB environment variable.

If the command is run without arguments and the either the --database global option or the
TMSU_DB environment variable is set, then the database will be created at this path instead,
with --database taking precedence.
"
    )]
    Init { path: Vec<PathBuf> },

    #[command(
        about = "List files with particular tags",
        alias = "query",
        long_about = "Lists files with particular tags.

QUERY is a space-separated list of tags that files must be tagged with in order to be listed. More
complex queries can also include operators and parentheses to further refine the results: see the
examples below.

Note: Queries match only tagged files. To identify untagged files use the 'untagged' subcommand.

Note: If your tag or value name contains whitespace, operators or parentheses, these must be escaped
with a backslash '\\', e.g. '\\<tag\\>' matches the tag name '<tag>'.

Note: Your shell may use some punctuation for its own purposes: this can usually be avoided by
enclosing the query in single quotation marks or by escaping the problematic characters with a
backslash.

Operators: and, or, not, ==, !=, <, >, <=, >=, eq, ne, lt, gt, le, ge.

Examples:

   $ tmsu files music mp3
   $ tmsu files music not mp3
   $ tmsu files 'music and (mp3 or flac)'
   $ tmsu files 'year == 2025'
   $ tmsu files 'year < 2020'
   $ tmsu files year lt 2020
   $ tmsu files --path=/some/path music
"
    )]
    Files {
        #[arg(help = "the query to run", num_args = 0..)]
        query: Vec<String>,
        #[arg(
            short = 'd',
            long = "directory",
            help = "list only items that are directories",
            default_value_t = false
        )]
        directory: bool,
        #[arg(
            short = 'f',
            long = "file",
            help = "list only items that are files",
            default_value_t = false
        )]
        file: bool,
        #[arg(
            short = 'c',
            long = "count",
            help = "list the number of matching files rather than their names",
            default_value_t = false
        )]
        count: bool,
        #[arg(short = 'p', long = "path", help = "list only items under PATH")]
        path: Option<PathBuf>,
        #[arg(
            short = 'e',
            long = "explicit",
            help = "list only explicitly tagged items",
            default_value_t = false
        )]
        explicit: bool,
        #[arg(
            short = 's',
            long = "sort",
            help = "sort output: id, name, none, size, time"
        )]
        sort: Option<String>,
        #[arg(
            short = 'i',
            long = "ignore-case",
            help = "ignore the case of tag and value names",
            default_value_t = false
        )]
        ignore_case: bool,
    },
}
