use crate::file::query::{
    Expression, Query,
    {And, Equal, GreaterOrEqual, GreaterThan, LessOrEqual, LessThan, Not, NotEqual, Or, Tagged},
};
use crate::tag::Tag;
use crate::value::Value;
use pest::{
    iterators::{Pair, Pairs},
    Parser,
};
use pest_derive::Parser;
use std::error::Error;

#[derive(Parser)]
#[grammar = "file/query/query.pest"]
pub struct QueryParser;

/// Parses the specified query text as an expression.
pub fn parse(text: &str) -> Result<Option<Query>, Box<dyn Error>> {
    let pairs = QueryParser::parse(Rule::query, text)?;
    let query = map_query(pairs);

    Ok(query)
}

fn map_query(mut parsed_query: Pairs<Rule>) -> Option<Query> {
    let expression = match parsed_query.next() {
        Some(pair) => map_pair(pair),
        None => None,
    };

    expression.map(|e| Query(e))
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
    let mut inner = pair.into_inner();
    let pair = inner.next().unwrap();

    match pair.as_rule() {
        Rule::unquoted_tag => map_unquoted_tag(pair).into(),
        Rule::quoted_tag => map_quoted_tag(pair).into(),
        _ => panic!("unexpected token: {}", pair.as_str()),
    }
}

fn map_unquoted_tag(pair: Pair<Rule>) -> Expression {
    let escaped = pair.as_str();
    let unescaped = escaped.replace("\\", "");

    Tagged(Tag(unescaped))
}

fn map_quoted_tag(pair: Pair<Rule>) -> Expression {
    let quoted_escaped = pair.as_str();
    let unescaped = quoted_escaped
        .chars()
        .skip(1)
        .take(quoted_escaped.len() - 2)
        .collect();

    Tagged(Tag(unescaped))
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
        let expected = Query(Tagged(Tag("single".into()).into()));

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_implicit_and() {
        let actual = parse("left right").unwrap().unwrap();
        let expected = Query(And(
            Tagged(Tag("left".into())).into(),
            Tagged(Tag("right".into())).into(),
        ));

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_explicit_and() {
        let actual = parse("left and right").unwrap().unwrap();
        let expected = Query(And(
            Tagged(Tag("left".into())).into(),
            Tagged(Tag("right".into())).into(),
        ));

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_quoted_tag() {
        let actual = parse("\"left and right\"").unwrap().unwrap();
        let expected = Query(Tagged(Tag("left and right".into()).into()));

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_comparisons() {
        let actual = parse("colour=red size == big wheels >= 4")
            .unwrap()
            .unwrap();
        let expected = Query(And(
            Equal(Tag("colour".into()), Value("red".into())).into(),
            And(
                Equal(Tag("size".into()), Value("big".into())).into(),
                GreaterOrEqual(Tag("wheels".into()), Value("4".into())).into(),
            )
            .into(),
        ));

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_operator_precedence() {
        let actual = parse("left or right and wrong").unwrap().unwrap();
        let expected = Query(Or(
            Tagged(Tag("left".into())).into(),
            And(
                Tagged(Tag("right".into())).into(),
                Tagged(Tag("wrong".into())).into(),
            )
            .into(),
        ));

        assert_eq!(expected, actual);
    }

    #[test]
    fn parse_parentheses() {
        let actual = parse("(left or right) and wrong").unwrap().unwrap();
        let expected = Query(And(
            Or(
                Tagged(Tag("left".into())).into(),
                Tagged(Tag("right".into())).into(),
            )
            .into(),
            Tagged(Tag("wrong".into())).into(),
        ));

        assert_eq!(expected, actual);
    }

    #[test]
    fn tag_names() {
        let actual = parse("colour == red and not (size == big or year < 2025)")
            .unwrap()
            .unwrap();

        assert_eq!(
            actual.tags(),
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
