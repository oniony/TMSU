use libtmsu::query;
use std::error::Error;
use std::path;
use std::path::PathBuf;

pub fn execute(
    _db_path: Option<PathBuf>,
    _verbosity: u8,
    query: Vec<String>,
    _directory: bool,
    _file: bool,
    _print0: bool,
    _count: bool,
    path: Option<PathBuf>,
    _explicit: bool,
    _sort: Option<String>,
    _ignore_case: bool,
) -> Result<(), Box<dyn Error>> {
    let _path = path.map(|p| path::absolute(p));
    let query = query::parse(query.join(" ").as_str())?;

    println!("query: {:?}", query);
    if let Some(query) = query {
        let tag_names = query.tags();
        let _value_names = query.values();

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
