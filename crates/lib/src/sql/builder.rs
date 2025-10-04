use rusqlite::types::ToSqlOutput;
use rusqlite::ToSql;
use std::error::Error;
use std::fmt::Display;

/// Allows composition of SQL statements with parameters.
pub struct SqlBuilder<'b> {
    sql: String,
    parameters: Vec<ToSqlOutput<'b>>,
}

impl<'b> SqlBuilder<'b> {
    pub fn new() -> SqlBuilder<'b> {
        SqlBuilder {
            sql: String::new(),
            parameters: Vec::new(),
        }
    }

    pub fn parameters(&self) -> &Vec<ToSqlOutput<'b>> {
        &self.parameters
    }

    pub fn sql(&mut self, sql: &str) -> &mut Self {
        if sql == "" {
            return self;
        }

        match &sql[0..1] {
            " " | "\n" => (),
            _ => self.sql.push('\n'),
        };

        self.sql.push_str(sql);

        self
    }

    pub fn parameter<T>(&mut self, param: &'b T) -> Result<&mut Self, Box<dyn Error>>
    where
        T: ToSql,
    {
        self.parameters.push(param.to_sql()?);
        let index = self.parameters.len();

        self.sql.push_str("?");
        self.sql.push_str(index.to_string().as_str());

        Ok(self)
    }
}

impl Display for SqlBuilder<'_> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.sql)
    }
}

