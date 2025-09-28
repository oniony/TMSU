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

use libtmsu::database::Database;
use libtmsu::query;
use std::error::Error;
use std::path;
use std::path::PathBuf;

pub fn execute(
    db_path: Option<PathBuf>,
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
    let db_path = db_path.ok_or("no database found")?;
    let _path = path.map(|p| path::absolute(p));
    let query = query::parse(query.join(" ").as_str())?;

    let database = Database::open(&db_path)?;

    if let Some(query) = query {
        let tags = query.tags();
        for invalid_tag in database.invalid_tags(&tags)? {
            eprintln!("unknown tag: {invalid_tag}")
        }

        let values = query.values();
        for invalid_value in database.invalid_values(&values)? {
            eprintln!("unknown value: {invalid_value}")
        }
    }

    //TODO run query
    //TODO handle parser stack overflow
    //TODO list the files

    println!("not implemented");

    Ok(())
}
