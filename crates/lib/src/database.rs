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

pub mod file;
pub mod query;
pub mod setting;
pub mod tag;
pub mod value;

use crate::database::setting::Setting;
use crate::migrations;
use rusqlite::Connection;
use std::error::Error;
use std::fs;
use std::path::{Path, PathBuf};

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

        let settings = setting::Store::new(&mut tx);
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

        let settings = setting::Store::new(&connection);
        let root_setting: PathBuf = settings.read(Setting::Root)?.into();

        let root = path
            .parent()
            .unwrap_or(&PathBuf::new())
            .join(root_setting);

        Ok(Database {
            path,
            root,
            connection: Some(connection),
        })
    }

    /// Retrieves the file store.
    pub fn files(&self) -> file::Store {
        file::Store::new(&self.connection.as_ref().unwrap())
    }

    /// Retrieves the tag store.
    pub fn tags(&self) -> tag::Store {
        tag::Store::new(&self.connection.as_ref().unwrap())
    }

    /// Retrieves the value store.
    pub fn values(&self) -> value::Store {
        value::Store::new(&self.connection.as_ref().unwrap())
    }
}

impl Drop for Database {
    fn drop(&mut self) {
        if let Some(connection) = self.connection.take() {
            let _ = connection.close();
        }
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::env;
    use std::fs::OpenOptions;

    #[test]
    fn properties() {
        let database = Database { path: PathBuf::from("some-path"), root: PathBuf::from("some-root"), connection: None };
        assert_eq!("some-path", database.path().to_str().unwrap());
        assert_eq!("some-root", database.root().to_str().unwrap());
    }

    #[test]
    fn create() {
        let path = env::temp_dir().join("tmsu-test");
        let _ = fs::remove_file(&path);
        let root = PathBuf::from("/some/root");

        Database::create(&path, &root).unwrap();
        assert!(path.exists());
    }

    #[test]
    fn create_collision() {
        let path = env::temp_dir().join("tmsu-test");
        let root = PathBuf::from("/some/root");

        OpenOptions::new()
            .create(true)
            .write(true)
            .open(&path)
            .unwrap();

        let error = Database::create(&path, &root);

        assert!(match error {
            Err(ref e) if e.to_string().contains("database already exists") => true,
            _ => false
        });
    }

    #[test]
    fn test_open() {
        let path = env::temp_dir().join("tmsu-test");
        let root = PathBuf::from("/some/root");
        let _ = fs::remove_file(&path);
        Database::create(&path, &root).unwrap();

        let database = Database::open(&path).unwrap();
        assert_eq!(path, database.path());
        assert_eq!(root, database.root());
    }
}
