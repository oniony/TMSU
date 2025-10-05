use crate::query::Expression;
use chrono::{DateTime, Utc};
use query::QueryBuilder;
use rusqlite::{params_from_iter, Connection, Rows};
use std::error::Error;
use std::path::PathBuf;

mod query;

/// File in the database.
pub struct File {
    _id: i64,
    directory: String,
    name: String,
    _fingerprint: String,
    _mod_time: DateTime<Utc>,
    _size: i64,
    _is_dir: bool,
}

impl File {
    pub fn path(&self) -> PathBuf {
        PathBuf::from(&self.directory).join(&self.name)
    }
}

/// The file store.
pub struct Store<'s> {
    connection: &'s Connection,
}

impl Store<'_> {
    /// Creates a new tag store.
    pub fn new(connection: &Connection) -> Store {
        Store { connection }
    }

    /// Queries for files by expression.
    pub fn query(&self, query: &Expression, explicit_only: bool, ignore_case: bool) -> Result<Vec<File>, Box<dyn Error>> {
        let mut builder = QueryBuilder::new(explicit_only, ignore_case);
        let (sql, parameters) = builder.file_query(&query)?;

        let mut statement = self.connection.prepare(&sql)?;
        let mut rows = statement.query(params_from_iter(parameters.iter()))?;

        Self::files_from_rows(&mut rows)
    }

    /// Queries the file count by expression.
    pub fn query_count(&self, query: &Expression, explicit_only: bool, ignore_case: bool) -> Result<u64, Box<dyn Error>> {
        let mut builder = QueryBuilder::new(explicit_only, ignore_case);
        let (sql, parameters) = builder.file_count_query(&query)?;

        let mut statement = self.connection.prepare(&sql)?;
        let count = statement.query_one(params_from_iter(parameters), |row| row.get::<usize, u64>(0))?;

        Ok(count)
    }

    /// Retrieves all files.
    pub fn all(&self) -> Result<Vec<File>, Box<dyn Error>> {
        let mut statement = self.connection.prepare("\
SELECT id, directory, name, fingerprint, mod_time, size, is_dir
FROM file;
")?;
        let mut rows = statement.query(())?;

        Self::files_from_rows(&mut rows)
    }

    pub fn all_count(&self) -> Result<u64, Box<dyn Error>> {
        let mut statement = self.connection.prepare("\
SELECT count(1)
FROM file;
")?;
        let count = statement.query_one((), |row| row.get::<usize, u64>(0))?;

        Ok(count)
    }

    fn files_from_rows(rows: &mut Rows) -> Result<Vec<File>, Box<dyn Error>> {
        let mut files = Vec::new();

        while let Some(row) = rows.next()? {
            let file = File {
                _id: row.get(0)?,
                directory: row.get(1)?,
                name: row.get(2)?,
                _fingerprint: row.get(3)?,
                _mod_time: row.get(4)?,
                _size: row.get(5)?,
                _is_dir: row.get(6)?,
            };

            files.push(file);
        }

        Ok(files)
    }
}
