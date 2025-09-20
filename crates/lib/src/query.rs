use std::error::Error;
use pest::iterators::Pairs;
use pest_derive::Parser;
use pest::Parser;

#[derive(Debug, PartialEq)]
pub struct Query {}

#[derive(Parser)]
#[grammar = "grammars/query.pest"]
struct QueryParser;

pub fn parse_query(text: &str) -> Result<Option<Query>, Box<dyn Error>> {
    let query = QueryParser::parse(Rule::query, text)?;

    dump(query, 0);

    Ok(Some(Query{}))
}

fn dump(query: Pairs<Rule>, depth: usize) {
    for pair in query {
        println!("{} {}", "  ".repeat(depth), match pair.as_rule() {
            Rule::tag => format!("tag({})", pair.as_str()),
            Rule::value => format!("value({})", pair.as_str()),

            Rule::and => "and".into(),
            Rule::or => "or".into(),
            Rule::not => "not".into(),

            Rule::equal => "==".into(),
            Rule::not_equal => "!=".into(),
            Rule::greater_than => ">".into(),
            Rule::less_than => "<".into(),
            Rule::greater_or_equal => ">=".into(),
            Rule::less_or_equal => "<=".into(),

            Rule::EOI => "".into(),
            _ => format!("other({:?})", pair.as_rule()),
        });

        dump(pair.into_inner(), depth + 1);
    }
}