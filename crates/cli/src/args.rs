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

use std::path::PathBuf;
use clap::{Parser, Subcommand};

#[derive(Parser)]
#[command(about = "TMSU", version, long_about = None)]
#[command(disable_colored_help = true)]
pub struct Args {
    #[clap(short = 'D', long)]
    pub database: Option<PathBuf>,

    #[command(subcommand)]
    pub command: Commands,
}

impl Args {
    pub fn parse() -> Self {
        Parser::parse()
    }
}

#[derive(Subcommand, Debug)]
pub enum Commands {
    #[command(
        about = "Show database information",
        long_about = "Show database paths and metrics.")]
    Info,

    #[command(
        about = "Initialize a new database",
        long_about = "Initializes new, empty databases.

Typically this command is run without arguments to create a new, local database at .tmsu/db in the
current working directory. Such a database will be used automatically when TMSU is used at or
below this directory.

If PATH is specified, then a database will be created at this specific path instead. Such a database
will only be subsequently used when TMSU is run with an explicit database path via the --database
global option or the TMSU_DATABASE environment variable.

If the command is run without arguments and the either the --database global option or the
TMSU_DATABASE environment variable is set, then the database will be created at this path instead,
with --database taking precedence.
")]
    Init { path: Vec<PathBuf> },
}
