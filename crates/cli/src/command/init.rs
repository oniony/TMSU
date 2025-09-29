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
use std::error::Error;
use std::path::PathBuf;
use std::{env, io, path};

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

    let pairs: Vec<(PathBuf, Result<PathBuf, io::Error>)> = paths
        .iter()
        .map(|p| (p.clone(), determine_root(p)))
        .collect();

    for (path, root) in pairs {
        Database::create(&path, &root?)?
    }

    Ok(())
}

fn determine_root(path: &PathBuf) -> Result<PathBuf, io::Error> {
    let abs_path = path::absolute(path)?;

    // if the database is within a '.tmsu' directory, then root the database to this directory's relative parent
    if let Some(parent) = abs_path.parent() {
        if let Some(filename) = parent.file_name() {
            if filename == APPLICATION_DIRECTORY {
                if let Some(_) = parent.parent() {
                    return Ok("..".into());
                }
            }
        }
    }

    Ok(PathBuf::from(path::MAIN_SEPARATOR_STR))
}

#[cfg(test)]
mod tests {
    use crate::command::init::determine_root;
    use std::path::PathBuf;

    #[test]
    fn determines_root() {
        let tests = [
            (PathBuf::from("/some/path"), PathBuf::from("/")),
            (PathBuf::from("/some/path/.tmsu/db"), PathBuf::from("..")),
        ];

        for test in tests {
            let actual_root = determine_root(&test.0).unwrap();
            let expected_root = test.1;
            assert_eq!(expected_root, actual_root);
        }
    }
}
