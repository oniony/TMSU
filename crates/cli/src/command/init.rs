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

use crate::constants::*;
use libtmsu::database::Database;
use std::env;
use std::error::Error;
use std::path::PathBuf;

pub fn execute(db_path: Option<PathBuf>, paths: Vec<PathBuf>) -> Result<(), Box<dyn Error>> {
    let paths = if paths.len() > 0 {
        paths
    } else if let Some(path) = db_path {
        vec![path]
    } else {
        vec![env::current_dir()?
            .join(APPLICATION_DIRECTORY)
            .join(DEFAULT_DATABASE_NAME)]
    };

    for path in paths {
        Database::create(&path)?
    }

    Ok(())
}
