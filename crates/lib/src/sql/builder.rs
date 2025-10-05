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
    /// Creates a new SQL builder.
    pub fn new() -> SqlBuilder<'b> {
        SqlBuilder {
            sql: String::new(),
            parameters: Vec::new(),
        }
    }

    /// Retrieves the parameter values.
    pub fn parameters(&self) -> &Vec<ToSqlOutput<'b>> {
        &self.parameters
    }

    /// Pushes a SQL string.
    pub fn push_sql(&mut self, sql: &str) -> &mut Self {
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

    /// Pushes a parameter.
    pub fn push_parameter<T>(&mut self, param: &'b T) -> Result<&mut Self, Box<dyn Error>>
    where
        T: ToSql,
    {
        self.parameters.push(param.to_sql()?);
        let index = self.parameters.len();

        self.sql.push_str("?");
        self.sql.push_str(index.to_string().as_str());

        Ok(self)
    }

    /// Pushes a set of values.
    pub fn push_parameterised_values<T>(&mut self, params: &'b [T]) -> Result<&mut Self, Box<dyn Error>>
    where
        T: ToSql,
    {
        let mut comma = false;
        for param in params {
            if comma {
                self.sql.push_str(",");
            }

            self.parameters.push(param.to_sql()?);
            let index = self.parameters.len();

            self.push_sql(&format!("(?{index})"));

            comma = true;
        }

        Ok(self)
    }
}

impl Display for SqlBuilder<'_> {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.sql)
    }
}

