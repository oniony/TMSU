use rusqlite::types::{ToSqlOutput, Value};
use rusqlite::ToSql;
use std::error::Error;
use std::fmt::Display;
use ToSqlOutput::Owned;

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
    pub fn parameters(self) -> Vec<ToSqlOutput<'b>> {
        self.parameters
    }

    /// Pushes a SQL string.
    pub fn push_sql(&mut self, sql: &str) -> &mut Self {
        if sql == "" {
            return self;
        }

        // add whitespace only if necessary
        if !self.sql.is_empty()
            && !self.sql.chars().last().unwrap().is_whitespace()
            && !sql.chars().next().unwrap().is_whitespace()
        {
            self.sql.push(' ');
        }

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

    /// Pushes a string parameter.
    pub fn push_parameter_string(&mut self, param: String) -> Result<&mut Self, Box<dyn Error>> {
        self.parameters.push(Owned(Value::Text(param)));

        let index = self.parameters.len();
        self.sql.push_str("?");
        self.sql.push_str(index.to_string().as_str());

        Ok(self)
    }

    /// Pushes a set of values.
    pub fn push_parameterised_values<T>(
        &mut self,
        params: &'b [T],
    ) -> Result<&mut Self, Box<dyn Error>>
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

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_sql_builder() {
        let mut builder = SqlBuilder::new();

        let values = vec!["hello", "there"];

        builder
            .push_sql("SELECT (")
            .push_parameterised_values(&values)
            .unwrap()
            .push_sql(") FROM users")
            .push_sql("WHERE id = ")
            .push_parameter(&1)
            .unwrap();

        assert_eq!(
            "SELECT ( (?1), (?2) ) FROM users WHERE id = ?3",
            builder.sql
        );

        let expected_parameters: Vec<ToSqlOutput> = vec!["hello".into(), "there".into(), 1.into()];
        assert_eq!(expected_parameters, builder.parameters)
    }
}
