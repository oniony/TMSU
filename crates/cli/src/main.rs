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

mod args;
mod command;
mod constants;
mod database;
mod error;
mod rendering;

use crate::error::MultiError;
use args::{Args, Commands};
use std::error::Error;
use std::process;

fn main() {
    let result = run();

    if let Err(error) = result {
        if let Some(validation_error) = error.downcast_ref::<MultiError>() {
            for error in &validation_error.errors {
                eprintln!("{}: {}", constants::APPLICATION_NAME, error);
            }
        } else {
            eprintln!("{}: {}", constants::APPLICATION_NAME, error);
        }

        process::exit(1)
    }
}

fn run() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let db_path = database::resolve(&args.database)?;
    let separator = args.separator();

    match args.command {
        Commands::Files {
            query,
            directory,
            file,
            count,
            path,
            explicit,
            sort,
            ignore_case,
        } => command::files::execute(
            database::open(db_path)?,
            args.verbosity,
            query,
            directory,
            file,
            separator,
            count,
            path,
            explicit,
            sort,
            ignore_case,
        ),
        Commands::Info => command::info::execute(database::open(db_path)?, separator),
        Commands::Init { path } => command::init::execute(db_path, path),
    }
}
