use std::error::Error;
use std::{fs, path};
use std::path::PathBuf;
use libtmsu::query;

pub fn execute(
    db_path: Option<PathBuf>,
    query: Vec<String>,
    directory: bool,
    file: bool,
    print0: bool,
    count: bool,
    path: Option<PathBuf>,
    explicit: bool,
    sort: Option<String>,
    ignore_case: bool,
) -> Result<(), Box<dyn Error>> {
    let path = path.map(|p| path::absolute(p));

    let query_text = query.join(" ");
    let query = query::parse_query(query_text.as_str())?;
    //TODO validate query tags
    //TODO validate query values
    //TODO run query
    //TODO handle parser stack overflow
    //TODO list the files

    println!("not implemented");

    Ok(())
}
