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
use crate::rendering::Separator;
use libtmsu::database::file::File;
use libtmsu::database::Database;
use libtmsu::query;
use libtmsu::query::Expression;
use std::error::Error;
use std::path;
use std::path::PathBuf;

/// Executes the 'files' command, which allows files to be queried by tag.
pub fn execute(
    database: Database,
    _verbosity: u8,
    args: Vec<String>,
    _directory: bool,
    _file: bool,
    separator: Separator,
    count: bool,
    path: Option<PathBuf>,
    explicit_only: bool,
    _sort: Option<String>,
    ignore_case: bool,
) -> Result<(), Box<dyn Error>> {
    let _path = path.map(|p| path::absolute(p));
    let query_text = args.join(" ").to_owned();
    let expression = query::parse(&query_text)?;

    let files = if let Some(expression) = &expression {
        query(database, expression, explicit_only, ignore_case)?
    } else {
        all(database)?
    };

    if count {
        show_count(&files, separator);
    } else {
        list_files(&files, separator);
    }

    Ok(())
}

fn query(
    database: Database,
    expression: &Expression,
    explicit_only: bool,
    ignore_case: bool)
    -> Result<Vec<File>, Box<dyn Error>> {
    let mut errors: Vec<Box<dyn Error + Send + Sync>> = Vec::new();

    let tags = expression.tags();
    let invalid_tags = database.tags().missing(&tags)?;
    for invalid_tag in &invalid_tags {
        errors.push(format!("unknown tag: {invalid_tag}").into());
    }

    let values = expression.values();
    let invalid_values = database.values().missing(&values)?;
    for invalid_value in &invalid_values {
        errors.push(format!("unknown value: {invalid_value}").into());
    }

    if !invalid_tags.is_empty() || !invalid_values.is_empty() {
        return Err(MultiError { errors }.into());
    }

    database.files().query(expression, explicit_only, ignore_case)
}

fn all(database: Database) -> Result<Vec<File>, Box<dyn Error>> {
    database.files().all()
}

fn show_count(files: &Vec<File>, separator: Separator) {
    print!("{}{separator}", files.len())
}

fn list_files(files: &Vec<File>, separator: Separator) {
    for file in files {
        print!("{}{separator}", file.path().to_str().unwrap_or(""));
    }
}
