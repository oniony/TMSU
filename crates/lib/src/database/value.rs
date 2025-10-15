use crate::sql::builder::SqlBuilder;
use rusqlite::{params_from_iter, Connection};
use std::error::Error;
use crate::database::common::Casing;
use crate::query::TagValue;

pub struct Store<'s> {
    connection: &'s Connection,
}

impl Store<'_> {
    /// Creates a new value store.
    pub fn new(connection: &Connection) -> Store {
        Store { connection }
    }

    /// Compares the specified set of values against the database, returning the set of missing value names.
    pub fn missing(&self, names: &[TagValue], casing: &Casing) -> Result<Vec<TagValue>, Box<dyn Error>> {
        if names.is_empty() {
            return Ok(vec![]);
        }

        let collation = match casing {
            Casing::Insensitive => "COLLATE NOCASE",
            Casing::Sensitive => "",
        };

        let mut builder = SqlBuilder::new();
        builder
            .push_sql("\
SELECT c.column1
FROM (
    VALUES ")
            .push_parameterised_values(names)?
            .push_sql(&format!("\
) AS c
LEFT JOIN value v ON c.column1 {collation} = v.name
WHERE v.name IS NULL;
"));

        let mut statement = self.connection.prepare(&builder.to_string())?;

        let invalid = statement
            .query_map(params_from_iter(builder.parameters()), |row| row.get::<usize, TagValue>(0))?
            .into_iter()
            .collect::<Result<Vec<_>, rusqlite::Error>>()?;

        Ok(invalid)
    }
}
