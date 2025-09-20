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

use std::error::Error;
use std::fs;
use std::path::PathBuf;
use rusqlite::Connection;
use crate::migrations;

pub fn create(path: &PathBuf) -> Result<(), Box<dyn Error>> {
    if path.exists() {
        return Err(format!("{}: Database already exists", path.to_str().unwrap()).into())
    }

    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)?;
    }

    let mut conn = Connection::open(path)?;
    let mut tx = conn.transaction()?;

    migrations::run(&mut tx)?;
    tx.commit()?;

    Ok(())
}
