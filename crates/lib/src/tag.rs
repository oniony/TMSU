use crate::common::Casing;
use crate::query::Tag;
use crate::sql::builder::SqlBuilder;
use rusqlite::{params_from_iter, Connection};
use std::error::Error;

/// The tag store.
pub struct Store<'s> {
    connection: &'s Connection,
}

impl Store<'_> {
    /// Creates a new tag store.
    pub fn new(connection: &Connection) -> Store {
        Store { connection }
    }

    /// Compares the specified set of tags against the database, returning the set of missing tag names.
    pub fn missing(&self, names: &[Tag], casing: &Casing) -> Result<Vec<Tag>, Box<dyn Error>> {
        if names.is_empty() {
            return Ok(vec![]);
        }

        let collation = match casing {
            Casing::Insensitive => "COLLATE NOCASE",
            Casing::Sensitive => "",
        };

        let mut builder = SqlBuilder::new();
        builder
            .push_sql(
                "\
SELECT c.column1
FROM (
    VALUES ",
            )
            .push_parameterised_values(names)?
            .push_sql(&format!(
                "\
) AS c
LEFT JOIN tag t ON c.column1 {collation} = t.name
WHERE t.name IS NULL;
        "
            ));

        let mut statement = self.connection.prepare(&builder.to_string())?;

        let invalid = statement
            .query_map(params_from_iter(builder.parameters()), |row| {
                row.get::<usize, Tag>(0)
            })?
            .into_iter()
            .collect::<Result<Vec<_>, rusqlite::Error>>()?;

        Ok(invalid)
    }
}
