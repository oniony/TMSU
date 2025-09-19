use std::error::Error;
use std::{fs, path};
use std::path::PathBuf;

pub fn execute(
    db_path: Option<PathBuf>,
    query: Option<String>,
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

    //TODO parse query (pest? nom?)
    //TODO validate query tags
    //TODO validate query values
    //TODO run query
    //TODO handle parser stack overflow
    //TODO list the files

    println!("not implemented");

    Ok(())
}
