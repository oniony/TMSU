use crate::query::Expression::{And, Equal, GreaterOrEqual, GreaterThan, LessOrEqual, LessThan, Not, NotEqual, Or, Tag};
use crate::query::{Expression, TagName, TagValue};
use crate::sql::builder::SqlBuilder;
use rusqlite::types::ToSqlOutput;
use std::error::Error;
use crate::database::common::{Casing, FileTypeSpecificity, TagSpecificity};

/// Builds a SQL query from a query expression.
pub struct QueryBuilder<'q> {
    tag_specificity: TagSpecificity,
    file_type: FileTypeSpecificity,
    casing: Casing,
    builder: SqlBuilder<'q>,
}

impl<'q> QueryBuilder<'q> {
    /// Creates a new QueryBuilder.
    pub fn new(tag_specificity: &TagSpecificity, file_type: &FileTypeSpecificity, casing: &Casing) -> QueryBuilder<'q> {
        QueryBuilder {
            tag_specificity: tag_specificity.clone(),
            file_type: file_type.clone(),
            casing: casing.clone(),
            builder: SqlBuilder::new(),
        }
    }

    /// Builds a SQL file query from the specified expression.
    pub fn file_query(
        &mut self,
        query: &'q Expression,
    ) -> Result<(String, &Vec<ToSqlOutput>), Box<dyn Error>> {
        self
            .select()
            .expression(&query)?
            .file_type();

        Ok((self.builder.to_string(), self.builder.parameters()))
    }

    /// Builds a SQL file count query from the specified expression.
    pub fn file_count_query(
        &mut self,
        query: &'q Expression,
    ) -> Result<(String, &Vec<ToSqlOutput>), Box<dyn Error>> {
        self
            .select()
            .expression(&query)?;

        Ok((self.builder.to_string(), self.builder.parameters()))
    }

    fn select(&mut self) -> &mut Self {
        self
            .builder.push_sql("\
SELECT id, directory, name, fingerprint, mod_time, size, is_dir
FROM file
WHERE");

        self
    }

    fn expression(&mut self, query: &'q Expression) -> Result<&mut Self, Box<dyn Error>> {
        match query {
            And(left, right) => self.binary(left, "AND", right),
            Equal(tag, value) => self.compare(tag, "=", value),
            GreaterThan(tag, value) => self.compare(tag, ">", value),
            GreaterOrEqual(tag, value) => self.compare(tag, ">=", value),
            LessThan(tag, value) => self.compare(tag, "<", value),
            LessOrEqual(tag, value) => self.compare(tag, "<=", value),
            NotEqual(tag, value) => self.compare(tag, "!=", value),
            Not(operand) => self.unary("NOT", operand),
            Or(left, right) => self.binary(left, "OR", right),
            Tag(tag) => self.tag(tag),
        }
    }

    fn compare(&mut self, tag: &'q TagName, operator: &str, value: &'q TagValue) -> Result<&mut Self, Box<dyn Error>> {
        self.builder.push_sql(&format!("-- {tag} {operator} {value}"));

        match self.tag_specificity {
            TagSpecificity::ExplicitOnly => self.compare_explicit(tag, operator, value),
            TagSpecificity::All => self.compare_all(tag, operator, value),
        }
    }

    fn compare_explicit(&mut self, tag: &'q TagName, operator: &str, value: &'q TagValue) -> Result<&mut Self, Box<dyn Error>> {
        let collation = self.collation();

        let negation = if operator == "!=" {
            "NOT"
        } else {
            ""
        };

        let operator = if operator == "!=" {
            "="
        } else {
            operator
        };

        self.builder.push_sql(&format!("\
id {negation} IN(
    WITH ift (tag_id, value_id) AS
        (
            SELECT t.id, v.id
            FROM tag t, value v
            WHERE t.name {collation} = "))
            .push_parameter(tag)?
            .push_sql(&format!("\
            AND v.name {collation} {operator} "))
            .push_parameter(value)?
            .push_sql("\
        )

    SELECT file_id
    FROM file_tag
    INNER JOIN ift
    ON file_tag.tag_id = ift.tag_id
    AND file_tag.value_id = ift.value_id
)");

        Ok(self)
    }

    fn compare_all(&mut self, tag: &'q TagName, operator: &str, value: &'q TagValue) -> Result<&mut Self, Box<dyn Error>> {
        let collation = self.collation();

        let negation = if operator == "!=" {
            "NOT"
        } else {
            ""
        };

        let operator = if operator == "!=" {
            "="
        } else {
            operator
        };

        self.builder.push_sql(&format!("\
id {negation} IN (
    WITH RECURSIVE ift (tag_id, value_id) AS
        (
            SELECT t.id, v.id
            FROM tag t, value v
            WHERE t.name {collation} = "))
            .push_parameter(tag)?
            .push_sql(&format!("\
            AND v.name {collation} {operator} "))
            .push_parameter(value)?
            .push_sql("\
            UNION ALL
            SELECT i.tag_id, i.value_id
            FROM implication i, ift
            WHERE i.implied_tag_id = ift.tag_id
            AND (ift.value_id = 0 OR i.implied_value_id = ift.value_id)
        )

    SELECT file_id
    FROM file_tag
    INNER JOIN ift
    ON file_tag.tag_id = ift.tag_id
    AND (file_tag.value_id = ift.value_id OR ift.value_id = 0)
)");

        Ok(self)
    }

    fn tag(&mut self, tag: &'q TagName) -> Result<&mut Self, Box<dyn Error>> {
        match self.tag_specificity {
            TagSpecificity::ExplicitOnly => self.tag_explicit(tag),
            TagSpecificity::All => self.tag_all(tag),
        }
    }

    fn unary(&mut self, operator: &str, operand: &'q Expression) -> Result<&mut Self, Box<dyn Error>> {
        self.builder.push_sql(operator);
        self.expression(operand)?;

        Ok(self)
    }

    fn binary(&mut self, left: &'q Expression, operator: &str, right: &'q Expression) -> Result<&mut Self, Box<dyn Error>> {
        self.expression(left)?;
        self.builder.push_sql(operator);
        self.expression(right)?;

        Ok(self)
    }

    fn tag_explicit(&mut self, tag: &'q TagName) -> Result<&mut Self, Box<dyn Error>> {
        let collation = self.collation();

        self.builder.push_sql(&format!("\
id IN (
    SELECT file_id
    FROM file_tag
    WHERE tag_id = (
        SELECT id
        FROM tag
        WHERE name {collation} = "))
            .push_parameter(tag)?
            .push_sql("\
    )\
)");

        Ok(self)
    }

    fn tag_all(&mut self, tag: &'q TagName) -> Result<&mut Self, Box<dyn Error>> {
        let collation = self.collation();

        self.builder.push_sql(&format!("\
id IN (
    WITH RECURSIVE ift (tag_id, value_id) AS
        (
            SELECT t.id, 0
            FROM tag t
            WHERE t.name {collation} = "))
            .push_parameter(tag)?
            .push_sql("\
            UNION ALL
            SELECT i.tag_id, i.value_id
            FROM implication i, ift
            WHERE i.implied_tag_id = ift.tag_id
            AND (ift.value_id = 0 OR i.implied_value_id = ift.value_id)
        )

    SELECT file_id
    FROM file_tag
    INNER JOIN ift
    ON file_tag.tag_id = ift.tag_id
    AND (file_tag.value_id = ift.value_id OR ift.value_id = 0)
)
");

        Ok(self)
    }

    fn collation(&self) -> &'static str {
        match self.casing {
            Casing::Insensitive => "COLLATE NOCASE",
            Casing::Sensitive => "",
        }
    }

    fn file_type(&mut self) -> &mut Self {
        self.builder.push_sql(match self.file_type {
            FileTypeSpecificity::Any => "",
            FileTypeSpecificity::FileOnly => "AND NOT is_dir",
            FileTypeSpecificity::DirectoryOnly => "AND is_dir",
        });

        self
    }
}
