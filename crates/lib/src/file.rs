use crate::common::Casing;
use crate::error::MultiError;
use crate::tag::TagSpecificity;
use crate::{tag, value};
use chrono::{DateTime, Utc};
use query::Query;
use rusqlite::{params_from_iter, Connection, Rows};
use std::error::Error;
use std::path::PathBuf;

pub(crate) mod query;

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

#[derive(Clone, Debug, Eq, PartialEq)]
pub enum FileTypeSpecificity {
    Any,
    FileOnly,
    DirectoryOnly,
}

/// The file store.
pub struct Store<'s> {
    connection: &'s Connection,
}

impl Store<'_> {
    /// Creates a new tag store.
    pub fn new(connection: &Connection) -> Store<'_> {
        Store { connection }
    }

    /// Queries for files by expression.
    pub fn query(
        &self,
        query_text: &str,
        tag_specificity: &TagSpecificity,
        file_type: &FileTypeSpecificity,
        casing: &Casing,
    ) -> Result<Vec<File>, Box<dyn Error>> {
        let query = query::parse(query_text)?;

        if let Some(query) = query {
            self.validate_query(&query, casing)?;

            let (sql, parameters) = query::files_sql(&query, tag_specificity, file_type, casing)?;
            let mut statement = self.connection.prepare(&sql)?;
            let mut rows = statement.query(params_from_iter(parameters.iter()))?;

            Self::files_from_rows(&mut rows)
        } else {
            self.all()
        }
    }

    /// Queries the file count by expression.
    pub fn query_count(
        &self,
        query_text: &str,
        tag_specificity: &TagSpecificity,
        file_type: &FileTypeSpecificity,
        casing: &Casing,
    ) -> Result<u64, Box<dyn Error>> {
        let query = query::parse(query_text)?;

        if let Some(query) = query {
            self.validate_query(&query, casing)?;

            let (sql, parameters) =
                query::file_count_sql(&query, tag_specificity, file_type, casing)?;
            let mut statement = self.connection.prepare(&sql)?;
            let count = statement
                .query_one(params_from_iter(parameters), |row| row.get::<usize, u64>(0))?;

            Ok(count)
        } else {
            self.all_count()
        }
    }

    /// Retrieves all files.
    pub fn all(&self) -> Result<Vec<File>, Box<dyn Error>> {
        let mut statement = self.connection.prepare(
            "\
SELECT id, directory, name, fingerprint, mod_time, size, is_dir
FROM file;
",
        )?;
        let mut rows = statement.query(())?;

        Self::files_from_rows(&mut rows)
    }

    pub fn all_count(&self) -> Result<u64, Box<dyn Error>> {
        let mut statement = self.connection.prepare(
            "\
SELECT count(1)
FROM file;
",
        )?;
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

    fn validate_query(&self, query: &Query, casing: &Casing) -> Result<(), Box<dyn Error>> {
        let mut errors: Vec<Box<dyn Error + Send + Sync>> = Vec::new();
        let expression = &query.0;

        let tags = expression.tags();
        let invalid_tags = tag::Store::new(self.connection).missing(&tags, &casing)?;
        for invalid_tag in &invalid_tags {
            errors.push(format!("unknown tag: {invalid_tag}").into());
        }

        let values = expression.values();
        let invalid_values = value::Store::new(self.connection).missing(&values, &casing)?;
        for invalid_value in &invalid_values {
            errors.push(format!("unknown value: {invalid_value}").into());
        }

        if errors.is_empty() {
            Ok(())
        } else {
            Err(MultiError { errors }.into())
        }
    }
}
