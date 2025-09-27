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
use colored::Colorize;
use std::error::Error;
use std::path::PathBuf;
use std::{io, path};

pub fn execute(db_path: Option<PathBuf>) -> Result<(), Box<dyn Error>> {
    let db_path = db_path.ok_or("no database found")?;
    let root_path = determine_root(&db_path)?;

    println!("Database path: {}", db_path.display().to_string().green());
    println!("Root path: {}", root_path.display().to_string().green());

    //TODO open database
    //TODO gather stats

    Ok(())
}

fn determine_root(path: &PathBuf) -> Result<PathBuf, io::Error> {
    let abs_path = path::absolute(path)?;

    if let Some(parent) = abs_path.parent() {
        if let Some(filename) = parent.file_name() {
            if filename == constants::APPLICATION_DIRECTORY {
                return Ok(parent.to_path_buf());
            }
        }
    }

    Ok(PathBuf::from(path::MAIN_SEPARATOR_STR))
}
