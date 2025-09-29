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

use crate::migrations;
use crate::query::{TagName, TagValue};
use rusqlite::types::FromSql;
use rusqlite::{Connection, ToSql, params_from_iter};
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
    pub path: PathBuf,
    pub root: PathBuf,
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
        Self::update_setting(&mut tx, Setting::Root, root.to_str().unwrap())?;

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
        let root_setting: PathBuf = Self::read_setting(&connection, Setting::Root)?.into();

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

    /// Validates the specified tags against the database, returning any that are invalid.
    pub fn invalid_tags(&self, tag_names: &[TagName]) -> Result<Vec<TagName>, Box<dyn Error>> {
        self.invalid_wotsits("tag", tag_names)
    }

    /// Validates the specified values against the database, returning any that are invalid.
    pub fn invalid_values(
        &self,
        value_names: &[TagValue],
    ) -> Result<Vec<TagValue>, Box<dyn Error>> {
        self.invalid_wotsits("value", value_names)
    }

    fn connection(&self) -> Result<&Connection, Box<dyn Error>> {
        self.connection
            .as_ref()
            .ok_or("database connection closed".into())
    }

    fn invalid_wotsits<N>(&self, wotsit: &str, names: &[N]) -> Result<Vec<N>, Box<dyn Error>>
    where
        N: ToSql + FromSql,
    {
        if names.is_empty() {
            return Ok(vec![]);
        }

        let connection = self.connection()?;

        let mut statement = connection.prepare(&format!(
            "\
            SELECT c.column1
            FROM (VALUES{}) AS c
            LEFT JOIN {wotsit} v ON c.column1 = v.name
            WHERE v.name IS NULL;",
            custom_placeholder_string("(?)", names.len())
        ))?;

        let invalid_names = statement
            .query_map(params_from_iter(names), |row| row.get::<usize, N>(0))?
            .into_iter()
            .collect::<Result<Vec<_>, rusqlite::Error>>()?;

        Ok(invalid_names)
    }

    fn read_setting(connection: &Connection, setting: Setting) -> Result<String, Box<dyn Error>> {
        let value = connection.query_one(
            "\
            SELECT value FROM setting WHERE name = ?;",
            [setting.to_string()],
            |r| r.get::<usize, String>(0),
        )?;

        Ok(value)
    }

    fn update_setting(
        connection: &Connection,
        setting: Setting,
        value: &str,
    ) -> Result<(), Box<dyn Error>> {
        let _ = connection.execute(
            "\
        INSERT INTO setting (name ,value)
        VALUES (?1, ?2)
        ON CONFLICT DO UPDATE
        SET value = ?2;
        ",
            (setting.to_string(), value),
        )?;

        Ok(())
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
