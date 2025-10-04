use crate::database::custom_placeholder_string;
use crate::query::TagName;
use rusqlite::{params_from_iter, Connection};
use std::error::Error;

pub struct Store<'s> {
    connection: &'s Connection,
}

impl Store<'_> {
    /// Creates a new tag store.
    pub fn new(connection: &Connection) -> Store {
        Store { connection }
    }

    /// Compares the specified set of tags against the database, returning the set of missing tag names.
    pub fn missing(&self, names: &[TagName]) -> Result<Vec<TagName>, Box<dyn Error>> {
        if names.is_empty() {
            return Ok(vec![]);
        }

        let mut statement = self.connection.prepare(&format!(
            "\
            SELECT c.column1
            FROM (VALUES{}) AS c
            LEFT JOIN tag t ON c.column1 = t.name
            WHERE t.name IS NULL;",
            custom_placeholder_string("(?)", names.len())
        ))?;

        let invalid = statement
            .query_map(params_from_iter(names), |row| row.get::<usize, TagName>(0))?
            .into_iter()
            .collect::<Result<Vec<_>, rusqlite::Error>>()?;

        Ok(invalid)
    }
}