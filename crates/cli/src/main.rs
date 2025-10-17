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
mod rendering;

use crate::args::FileType;
use crate::command::files::FilesCommand;
use crate::command::info::InfoCommand;
use crate::command::init::InitCommand;
use args::{Args, Commands};
use libtmsu::common::Casing;
use libtmsu::error::MultiError;
use libtmsu::file::FileTypeSpecificity;
use libtmsu::tag::TagSpecificity;
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

trait Executor {
    fn execute(&self) -> Result<(), Box<dyn Error>>;
}

fn run() -> Result<(), Box<dyn Error>> {
    let args = Args::parse();
    let db_path = database::resolve(&args.database)?;
    let separator = args.separator();

    let executor: &dyn Executor = match args.command {
        Commands::Files {
            count,
            explicit,
            file_type,
            ignore_case,
            query,
        } => &FilesCommand::new(
            database::open(db_path)?,
            query,
            separator,
            count,
            if explicit {
                TagSpecificity::ExplicitOnly
            } else {
                TagSpecificity::All
            },
            match file_type {
                FileType::Any => FileTypeSpecificity::Any,
                FileType::FileOnly => FileTypeSpecificity::FileOnly,
                FileType::DirectoryOnly => FileTypeSpecificity::DirectoryOnly,
            },
            if ignore_case {
                Casing::Insensitive
            } else {
                Casing::Sensitive
            },
        ),
        Commands::Info => &InfoCommand::new(database::open(db_path)?, separator),
        Commands::Init { paths } => &InitCommand::new(db_path, paths),
    };

    executor.execute()
}
