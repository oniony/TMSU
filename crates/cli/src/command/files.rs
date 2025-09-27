use libtmsu::query;
use std::error::Error;
use std::path::PathBuf;
use std::{fs, path};

pub fn execute(
    db_path: Option<PathBuf>,
    verbosity: u8,
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
    let query = query::parse(query.join(" ").as_str())?;

    println!("query: {:?}", query);
    if let Some(query) = query {
        let tag_names = query.tags();
        let value_names = query.values();

        println!("tag names: {:?}", tag_names);
        //TODO validate query tags
        //TODO validate query values
    }

    //TODO run query
    //TODO handle parser stack overflow
    //TODO list the files

    println!("not implemented");

    Ok(())
}
