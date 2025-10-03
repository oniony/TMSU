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

use crate::error::MultiError;
use libtmsu::database::Database;
use libtmsu::query;
use std::error::Error;
use std::path;
use std::path::PathBuf;

/// Executes the 'files' command, which allows files to be queried by tag.
pub fn execute(
    database: Database,
    _verbosity: u8,
    query: Vec<String>,
    _directory: bool,
    _file: bool,
    _print0: bool,
    _count: bool,
    path: Option<PathBuf>,
    _explicit: bool,
    _sort: Option<String>,
    _ignore_case: bool,
) -> Result<(), Box<dyn Error>> {
    let _path = path.map(|p| path::absolute(p));
    let query = query::parse(query.join(" ").as_str())?;

    let mut errors: Vec<Box<dyn Error + Send + Sync>> = Vec::new();

    if let Some(query) = query {
        let tags = query.tags();
        for invalid_tag in database.tags().missing(&tags)? {
            errors.push(format!("unknown tag: {invalid_tag}").into());
        }

        let values = query.values();
        for invalid_value in database.values().missing(&values)? {
            errors.push(format!("unknown value: {invalid_value}").into());
        }
    }

    if !errors.is_empty() {
        return Err(MultiError { errors }.into());
    }

    //TODO run query
    //TODO handle parser stack overflow
    //TODO list the files

    println!("not implemented");

    Ok(())
}
