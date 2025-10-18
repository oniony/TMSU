// Copyright 2011-2025 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

mod builder;
mod parser;

use crate::common::Casing;
use crate::file::query::builder::QueryBuilder;
use crate::file::query::Expression::*;
use crate::file::FileTypeSpecificity;
use crate::tag::{Tag, TagSpecificity};
use crate::value::Value;
use rusqlite::types::ToSqlOutput;
use std::error::Error;

/// Parse a textual query.
pub fn parse(text: &str) -> Result<Option<Query>, Box<dyn Error>> {
    parser::parse(text)
}

/// Builds SQL for a files query.
pub fn files_sql<'q>(
    query: &'q Query,
    tag_specificity: &'q TagSpecificity,
    file_type_specificity: &'q FileTypeSpecificity,
    casing: &'q Casing,
) -> Result<(String, Vec<ToSqlOutput<'q>>), Box<dyn Error>> {
    let qb = QueryBuilder::new(&tag_specificity, &file_type_specificity, &casing);
    let sql_and_params = qb.file_query(query)?;

    Ok(sql_and_params)
}

/// Builds SQL for a file count query.
pub fn file_count_sql<'q>(
    query: &'q Query,
    tag_specificity: &'q TagSpecificity,
    file_type_specificity: &'q FileTypeSpecificity,
    casing: &'q Casing,
) -> Result<(String, Vec<ToSqlOutput<'q>>), Box<dyn Error>> {
    let qb = QueryBuilder::new(&tag_specificity, &file_type_specificity, &casing);
    let sql_and_params = qb.file_count_query(query)?;

    Ok(sql_and_params)
}

/// A parsed query.
#[derive(Debug, PartialEq, Eq)]
pub struct Query(Expression);

impl Query {
    /// Identifies the tags within the query.
    pub fn tags(&self) -> Vec<Tag> {
        self.0.tags()
    }

    /// Identifies the values within the query.
    pub fn values(&self) -> Vec<Value> {
        self.0.values()
    }
}

/// A query expression.
#[derive(Debug, PartialEq, Eq)]
pub enum Expression {
    Tagged(Tag),
    And(Box<Expression>, Box<Expression>),
    Or(Box<Expression>, Box<Expression>),
    Not(Box<Expression>),
    Equal(Tag, Value),
    NotEqual(Tag, Value),
    GreaterThan(Tag, Value),
    LessThan(Tag, Value),
    GreaterOrEqual(Tag, Value),
    LessOrEqual(Tag, Value),
}

impl Expression {
    /// Identifies the tag names within the expression.
    pub fn tags(&self) -> Vec<Tag> {
        let mut tag_names = Vec::new();
        self.walk_tags(&mut tag_names);
        tag_names
    }

    /// Identifies the value names within the expression.
    pub fn values(&self) -> Vec<Value> {
        let mut value_names = Vec::new();
        self.walk_values(&mut value_names);
        value_names
    }

    fn walk_tags(&self, tags: &mut Vec<Tag>) {
        match &self {
            Tagged(tag) => tags.push(tag.clone()),
            And(left, right) | Or(left, right) => {
                left.walk_tags(tags);
                right.walk_tags(tags);
            }
            Not(operand) => operand.walk_tags(tags),
            Equal(tag_name, _)
            | NotEqual(tag_name, _)
            | GreaterThan(tag_name, _)
            | LessThan(tag_name, _)
            | GreaterOrEqual(tag_name, _)
            | LessOrEqual(tag_name, _) => tags.push(tag_name.clone()),
        }
    }

    fn walk_values(&self, values: &mut Vec<Value>) {
        match &self {
            And(left, right) | Or(left, right) => {
                left.walk_values(values);
                right.walk_values(values);
            }
            Not(operand) => operand.walk_values(values),
            Equal(_, value_name)
            | NotEqual(_, value_name)
            | GreaterThan(_, value_name)
            | LessThan(_, value_name)
            | GreaterOrEqual(_, value_name)
            | LessOrEqual(_, value_name) => values.push(value_name.clone()),
            _ => (),
        }
    }
}
