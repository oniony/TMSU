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

use args::{Args, Commands};
use std::error::Error;
use std::process;

#[tokio::main]
async fn main() {
    let args = Args::parse();

    let db_path = match database::resolve(args.database) {
        Ok(db_path) => db_path,
        Err(error) => return fatal(error),
    };

    let result = match args.command {
        Commands::Files {
            query,
            directory,
            file,
            print0,
            count,
            path,
            explicit,
            sort,
            ignore_case,
        } => command::files::execute(
            db_path,
            args.verbosity,
            query,
            directory,
            file,
            print0,
            count,
            path,
            explicit,
            sort,
            ignore_case,
        ),
        Commands::Info => command::info::execute(db_path),
        Commands::Init { path } => command::init::execute(db_path, path),
    };

    match result {
        Ok(_) => (),
        Err(error) => fatal(error),
    }
}

fn fatal(error: Box<dyn Error>) {
    eprintln!("{}: {}", constants::APPLICATION_NAME, error.to_string());
    process::exit(1)
}
