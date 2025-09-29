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

use crate::constants;
use libtmsu::database::Database;
use std::env;
use std::env::VarError;
use std::error::Error;
use std::path::PathBuf;

/// Resolves the database path considering the program argument, environment variable and working directory.
pub fn resolve(arg_db_path: Option<PathBuf>) -> Result<Option<PathBuf>, Box<dyn Error>> {
    match arg_db_path {
        Some(path) => Ok(Some(path)),
        None => match env::var(constants::DATABASE_ENV_VAR) {
            Ok(path) => Ok(Some(PathBuf::from(path))),
            Err(VarError::NotPresent) => find(),
            Err(err) => Err(err.into()),
        },
    }
}

/// Opens the database at the specified path.
pub fn open(path: Option<PathBuf>) -> Result<Database, Box<dyn Error>> {
    let path = path.ok_or("no database found")?;

    Database::open(&path)
}

fn find() -> Result<Option<PathBuf>, Box<dyn Error>> {
    let working_dir = env::current_dir()?;
    let mut search = working_dir.clone();

    loop {
        let candidate = search
            .join(constants::APPLICATION_DIRECTORY)
            .join(constants::DEFAULT_DATABASE_NAME);

        if candidate.exists() {
            return Ok(Some(candidate));
        }

        if !search.pop() {
            return Ok(None);
        }
    }
}
