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
use rusqlite::{params_from_iter, Connection, ToSql};
use std::error::Error;
use std::fs;
use std::path::PathBuf;

// An application database.
pub struct Database {
    connection: Option<Connection>,
}

impl Database {
    /// Creates a new, empty database at the specified path.
    pub fn create(path: &PathBuf) -> Result<(), Box<dyn Error>> {
        if path.exists() {
            return Err(format!("{}: database already exists", path.to_str().unwrap()).into());
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

    /// Opens the database at the specified path.
    pub fn open(path: &PathBuf) -> Result<Database, Box<dyn Error>> {
        if !path.exists() {
            return Err(format!("{}: database not found", path.to_str().unwrap()).into());
        }

        let connection = Connection::open(path)?;

        Ok(Database {
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
