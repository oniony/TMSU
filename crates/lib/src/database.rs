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

mod settings;
mod tags;
mod values;

use crate::migrations;
use rusqlite::Connection;
use std::error::Error;
use std::fmt::Display;
use std::fs;
use std::path::{Path, PathBuf};

/// Application settings.
pub enum Setting {
    Root,
}

impl Display for Setting {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            Setting::Root => write!(f, "database-root"),
        }
    }
}

// An application database.
pub struct Database {
    path: PathBuf,
    root: PathBuf,
    connection: Option<Connection>,
}

impl Database {
    /// The database file path.
    pub fn path(&self) -> &Path {
        &self.path
    }

    /// The database root, from which file paths are relative.
    pub fn root(&self) -> &Path {
        &self.root
    }

    /// Creates a new, empty database at the specified path.
    pub fn create(path: &Path, root: &Path) -> Result<(), Box<dyn Error>> {
        if path.exists() {
            return Err(format!("{}: database already exists", path.to_str().unwrap()).into());
        }

        if let Some(parent) = path.parent() {
            fs::create_dir_all(parent)?;
        }

        let mut conn = Connection::open(path)?;
        let mut tx = conn.transaction()?;

        migrations::run(&mut tx)?;

        let settings = settings::Store::new(&mut tx);
        settings.update(Setting::Root, root.to_str().unwrap())?;

        tx.commit()?;

        Ok(())
    }

    /// Opens the database at the specified path.
    pub fn open(path: &PathBuf) -> Result<Database, Box<dyn Error>> {
        if !path.exists() {
            return Err(format!("{}: database not found", path.to_str().unwrap()).into());
        }

        let connection = Connection::open(path)?;
        let path = path.clone();

        let settings = settings::Store::new(&connection);
        let root_setting: PathBuf = settings.read(Setting::Root)?.into();

        let root = path
            .parent()
            .unwrap_or(&PathBuf::new())
            .join(root_setting)
            .canonicalize()?;

        Ok(Database {
            path,
            root,
            connection: Some(connection),
        })
    }

    /// Retrieves the tag store.
    pub fn tags(&self) -> tags::Store {
        tags::Store::new(&self.connection.as_ref().unwrap())
    }

    /// Retrieves the value store.
    pub fn values(&self) -> values::Store {
        values::Store::new(&self.connection.as_ref().unwrap())
    }
}

impl Drop for Database {
    fn drop(&mut self) {
        if let Some(connection) = self.connection.take() {
            let _ = connection.close();
        }
    }
}

#[inline]
fn custom_placeholder_string(placeholder: &str, count: usize) -> String {
    std::iter::repeat(placeholder)
        .take(count)
        .collect::<Vec<_>>()
        .join(", ")
}
