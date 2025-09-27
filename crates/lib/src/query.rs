use crate::query::Expression::*;
use pest::Parser;
use pest::iterators::{Pair, Pairs};
use pest_derive::Parser;
use std::error::Error;

/// Tag name.
#[derive(Debug, PartialEq, Eq, Clone)]
pub struct TagName(String);

// Tag value.
#[derive(Debug, PartialEq, Eq, Clone)]
pub struct TagValue(String);

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
    pub fn tags(&self) -> Vec<TagName> {
        let mut tag_names = Vec::new();
        self.walk_tag_names(&mut tag_names);
        tag_names
    }

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
    println!("parsing {text}");
    let parsed_query = QueryParser::parse(Rule::query, text)?;
    let query = map_query(parsed_query);

    Ok(query)
}

fn map_query(mut parsed_query: Pairs<Rule>) -> Option<Expression> {
    match parsed_query.next() {
        Some(pair) => Some(map_pair(pair)),
        None => None,
    }
}

fn map_pair(pair: Pair<Rule>) -> Expression {
    match pair.as_rule() {
        Rule::tag => map_tag(pair),
        Rule::equal => map_comparison_operator(pair, Equal),
        Rule::not_equal => map_comparison_operator(pair, NotEqual),
        Rule::greater_than => map_comparison_operator(pair, GreaterThan),
        Rule::less_than => map_comparison_operator(pair, LessThan),
        Rule::greater_or_equal => map_comparison_operator(pair, GreaterOrEqual),
        Rule::less_or_equal => map_comparison_operator(pair, LessOrEqual),
        Rule::and => map_binary_logical_operator(pair, And),
        Rule::or => map_binary_logical_operator(pair, Or),
        Rule::not => map_unary_logical_operator(pair, Not),
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
    let left_operand = map_pair(inner.next().unwrap());
    let right_operand = map_pair(inner.next().unwrap());

    factory(left_operand.into(), right_operand.into())
}

fn map_unary_logical_operator<F>(pair: Pair<Rule>, factory: F) -> Expression
where
    F: Fn(Box<Expression>) -> Expression,
{
    let mut inner = pair.into_inner();
    let operand = map_pair(inner.next().unwrap());

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
