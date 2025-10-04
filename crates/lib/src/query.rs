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

/// Tag name.
#[derive(Debug, PartialEq, Eq, Clone)]
pub struct TagName(String);

impl Display for TagName {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl ToSql for TagName {
    fn to_sql(&self) -> rusqlite::Result<ToSqlOutput<'_>> {
        Ok(ToSqlOutput::from(self.0.as_str()))
    }
}

impl FromSql for TagName {
    fn column_result(value: rusqlite::types::ValueRef<'_>) -> rusqlite::Result<Self, FromSqlError> {
        Ok(Self(value.as_str()?.to_string()))
    }
}

// Tag value.
#[derive(Debug, PartialEq, Eq, Clone)]
pub struct TagValue(String);

impl Display for TagValue {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.0)
    }
}

impl ToSql for TagValue {
    fn to_sql(&self) -> rusqlite::Result<ToSqlOutput<'_>> {
        Ok(ToSqlOutput::from(self.0.as_str()))
    }
}

impl FromSql for TagValue {
    fn column_result(value: rusqlite::types::ValueRef<'_>) -> rusqlite::Result<Self, FromSqlError> {
        Ok(Self(value.as_str()?.to_string()))
    }
}

/// A query expression.
#[derive(Debug, PartialEq, Eq)]
pub enum Expression {
    Tag(TagName),
    And(Box<Expression>, Box<Expression>),
    Or(Box<Expression>, Box<Expression>),
    Not(Box<Expression>),
    Equal(TagName, TagValue),
    NotEqual(TagName, TagValue),
    GreaterThan(TagName, TagValue),
    LessThan(TagName, TagValue),
    GreaterOrEqual(TagName, TagValue),
    LessOrEqual(TagName, TagValue),
}

impl Expression {
    /// Identifies the tag names within the expression.
    pub fn tags(&self) -> Vec<TagName> {
        let mut tag_names = Vec::new();
        self.walk_tag_names(&mut tag_names);
        tag_names
    }

    /// Identifies the value names within the expression.
    pub fn values(&self) -> Vec<TagValue> {
        let mut value_names = Vec::new();
        self.walk_value_names(&mut value_names);
        value_names
    }

    fn walk_tag_names(&self, tag_names: &mut Vec<TagName>) {
        match &self {
            Tag(tag_name) => tag_names.push(tag_name.clone()),
            And(left, right) | Or(left, right) => {
                left.walk_tag_names(tag_names);
                right.walk_tag_names(tag_names);
            }
            Not(operand) => operand.walk_tag_names(tag_names),
            Equal(tag_name, _)
            | NotEqual(tag_name, _)
            | GreaterThan(tag_name, _)
            | LessThan(tag_name, _)
            | GreaterOrEqual(tag_name, _)
            | LessOrEqual(tag_name, _) => tag_names.push(tag_name.clone()),
        }
    }

    fn walk_value_names(&self, value_names: &mut Vec<TagValue>) {
        match &self {
            And(left, right) | Or(left, right) => {
                left.walk_value_names(value_names);
                right.walk_value_names(value_names);
            }
            Not(operand) => operand.walk_value_names(value_names),
            Equal(_, value_name)
            | NotEqual(_, value_name)
            | GreaterThan(_, value_name)
            | LessThan(_, value_name)
            | GreaterOrEqual(_, value_name)
            | LessOrEqual(_, value_name) => value_names.push(value_name.clone()),
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
    Tag(TagName(pair.as_str().into()))
}

fn map_comparison_operator<F>(pair: Pair<Rule>, factory: F) -> Expression
where
    F: Fn(TagName, TagValue) -> Expression,
{
    let mut inner = pair.into_inner();
    let tag = TagName(inner.next().unwrap().as_str().into());
    let value = TagValue(inner.next().unwrap().as_str().into());

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
        let expected = Tag(TagName("single".into()).into());

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_implicit_and() {
        let actual = parse("left right").unwrap().unwrap();
        let expected = And(
            Tag(TagName("left".into())).into(),
            Tag(TagName("right".into())).into(),
        );

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_explicit_and() {
        let actual = parse("left and right").unwrap().unwrap();
        let expected = And(
            Tag(TagName("left".into())).into(),
            Tag(TagName("right".into())).into(),
        );

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_quoted_tag() {
        let actual = parse("\"left and right\"").unwrap().unwrap();
        let expected = Tag(TagName("\"left and right\"".into()).into());

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_comparisons() {
        let actual = parse("colour=red size == big wheels >= 4")
            .unwrap()
            .unwrap();
        let expected = And(
            Equal(TagName("colour".into()), TagValue("red".into())).into(),
            And(
                Equal(TagName("size".into()), TagValue("big".into())).into(),
                GreaterOrEqual(TagName("wheels".into()), TagValue("4".into())).into(),
            )
                .into(),
        );

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_operator_precedence() {
        let actual = parse("left or right and wrong").unwrap().unwrap();
        let expected = Or(
            Tag(TagName("left".into())).into(),
            And(
                Tag(TagName("right".into())).into(),
                Tag(TagName("wrong".into())).into(),
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
                Tag(TagName("left".into())).into(),
                Tag(TagName("right".into())).into(),
            )
                .into(),
            Tag(TagName("wrong".into())).into(),
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
            vec![
                TagName("colour".into()),
                TagName("size".into()),
                TagName("year".into())
            ]
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
                TagValue("red".into()),
                TagValue("big".into()),
                TagValue("2025".into())
            ]
        );
    }
}
