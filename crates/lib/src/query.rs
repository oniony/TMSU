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

use crate::query::Expression::*;
use pest::iterators::{Pair, Pairs};
use pest::Parser;
use pest_derive::Parser;
use rusqlite::types::{FromSql, FromSqlError, ToSqlOutput};
use rusqlite::ToSql;
use std::error::Error;
use std::fmt::Display;

/// Tag.
#[derive(Debug, PartialEq, Eq, Clone)]
pub struct Tag(String);

impl Display for Tag {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl ToSql for Tag {
    fn to_sql(&self) -> rusqlite::Result<ToSqlOutput<'_>> {
        Ok(ToSqlOutput::from(self.0.as_str()))
    }
}

impl FromSql for Tag {
    fn column_result(value: rusqlite::types::ValueRef<'_>) -> rusqlite::Result<Self, FromSqlError> {
        Ok(Self(value.as_str()?.to_string()))
    }
}

// Tag value.
#[derive(Debug, PartialEq, Eq, Clone)]
pub struct Value(String);

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

#[derive(Parser)]
#[grammar = "grammars/query.pest"]
struct QueryParser;

/// Parse a query string into an expression tree.
pub fn parse(text: &str) -> Result<Option<Expression>, Box<dyn Error>> {
    let parsed_query = QueryParser::parse(Rule::query, text)?;
    let query = map_query(parsed_query);

    Ok(query)
}

fn map_query(mut parsed_query: Pairs<Rule>) -> Option<Expression> {
    match parsed_query.next() {
        Some(pair) => map_pair(pair),
        None => None,
    }
}

fn map_pair(pair: Pair<Rule>) -> Option<Expression> {
    match pair.as_rule() {
        Rule::tag => map_tag(pair).into(),
        Rule::equal => map_comparison_operator(pair, Equal).into(),
        Rule::not_equal => map_comparison_operator(pair, NotEqual).into(),
        Rule::greater_than => map_comparison_operator(pair, GreaterThan).into(),
        Rule::less_than => map_comparison_operator(pair, LessThan).into(),
        Rule::greater_or_equal => map_comparison_operator(pair, GreaterOrEqual).into(),
        Rule::less_or_equal => map_comparison_operator(pair, LessOrEqual).into(),
        Rule::and => map_binary_logical_operator(pair, And).into(),
        Rule::or => map_binary_logical_operator(pair, Or).into(),
        Rule::not => map_unary_logical_operator(pair, Not).into(),
        Rule::EOI => None,
        _ => panic!("unexpected token: {}", pair.as_str()),
    }
}

fn map_tag(pair: Pair<Rule>) -> Expression {
    Tagged(Tag(pair.as_str().into()))
}

fn map_comparison_operator<F>(pair: Pair<Rule>, factory: F) -> Expression
where
    F: Fn(Tag, Value) -> Expression,
{
    let mut inner = pair.into_inner();
    let tag = Tag(inner.next().unwrap().as_str().into());
    let value = Value(inner.next().unwrap().as_str().into());

    factory(tag, value)
}

fn map_binary_logical_operator<F>(pair: Pair<Rule>, factory: F) -> Expression
where
    F: Fn(Box<Expression>, Box<Expression>) -> Expression,
{
    let mut inner = pair.into_inner();
    let left_operand = map_pair(inner.next().unwrap()).unwrap();
    let right_operand = map_pair(inner.next().unwrap()).unwrap();

    factory(left_operand.into(), right_operand.into())
}

fn map_unary_logical_operator<F>(pair: Pair<Rule>, factory: F) -> Expression
where
    F: Fn(Box<Expression>) -> Expression,
{
    let mut inner = pair.into_inner();
    let operand = map_pair(inner.next().unwrap()).unwrap();

    factory(operand.into())
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn parse_single_tag_query() {
        let actual = parse("single").unwrap().unwrap();
        let expected = Tagged(Tag("single".into()).into());

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_implicit_and() {
        let actual = parse("left right").unwrap().unwrap();
        let expected = And(
            Tagged(Tag("left".into())).into(),
            Tagged(Tag("right".into())).into(),
        );

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_explicit_and() {
        let actual = parse("left and right").unwrap().unwrap();
        let expected = And(
            Tagged(Tag("left".into())).into(),
            Tagged(Tag("right".into())).into(),
        );

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_quoted_tag() {
        let actual = parse("\"left and right\"").unwrap().unwrap();
        let expected = Tagged(Tag("\"left and right\"".into()).into());

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_comparisons() {
        let actual = parse("colour=red size == big wheels >= 4")
            .unwrap()
            .unwrap();
        let expected = And(
            Equal(Tag("colour".into()), Value("red".into())).into(),
            And(
                Equal(Tag("size".into()), Value("big".into())).into(),
                GreaterOrEqual(Tag("wheels".into()), Value("4".into())).into(),
            )
            .into(),
        );

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_operator_precedence() {
        let actual = parse("left or right and wrong").unwrap().unwrap();
        let expected = Or(
            Tagged(Tag("left".into())).into(),
            And(
                Tagged(Tag("right".into())).into(),
                Tagged(Tag("wrong".into())).into(),
            )
            .into(),
        );

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_parentheses() {
        let actual = parse("(left or right) and wrong").unwrap().unwrap();
        let expected = And(
            Or(
                Tagged(Tag("left".into())).into(),
                Tagged(Tag("right".into())).into(),
            )
            .into(),
            Tagged(Tag("wrong".into())).into(),
        );

        assert_eq!(expected, actual);
    }

    #[test]
    fn tag_names() {
        let query = parse("colour == red and not (size == big or year < 2025)")
            .unwrap()
            .unwrap();

        assert_eq!(
            query.tags(),
            vec![Tag("colour".into()), Tag("size".into()), Tag("year".into())]
        );
    }

    #[test]
    fn value_names() {
        let query = parse("colour == red and not (size == big or year < 2025)")
            .unwrap()
            .unwrap();

        assert_eq!(
            query.values(),
            vec![
                Value("red".into()),
                Value("big".into()),
                Value("2025".into())
            ]
        );
    }
}
