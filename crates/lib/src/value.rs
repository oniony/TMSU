use crate::common::Casing;
use crate::sql::builder::SqlBuilder;
use rusqlite::types::{FromSql, FromSqlError, ToSqlOutput};
use rusqlite::{params_from_iter, Connection, ToSql};
use std::error::Error;
use std::fmt::Display;

// Tag value.
#[derive(Debug, PartialEq, Eq, Clone)]
pub struct Value(pub String);

impl Display for Value {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl ToSql for Value {
    fn to_sql(&self) -> rusqlite::Result<ToSqlOutput<'_>> {
        Ok(ToSqlOutput::from(self.0.as_str()))
    }
}

impl FromSql for Value {
    fn column_result(value: rusqlite::types::ValueRef<'_>) -> rusqlite::Result<Self, FromSqlError> {
        Ok(Self(value.as_str()?.to_string()))
    }
}

pub struct Store<'s> {
    connection: &'s Connection,
}

impl Store<'_> {
    /// Creates a new value store.
    pub fn new(connection: &Connection) -> Store<'_> {
        Store { connection }
    }

    /// Compares the specified set of values against the database, returning the set of missing value names.
    pub fn missing(&self, names: &[Value], casing: &Casing) -> Result<Vec<Value>, Box<dyn Error>> {
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
LEFT JOIN value v ON c.column1 {collation} = v.name
WHERE v.name IS NULL;
"
            ));

        let mut statement = self.connection.prepare(&builder.to_string())?;

        let invalid = statement
            .query_map(params_from_iter(builder.parameters()), |row| {
                row.get::<usize, Value>(0)
            })?
            .into_iter()
            .collect::<Result<Vec<_>, rusqlite::Error>>()?;

        Ok(invalid)
    }
}
