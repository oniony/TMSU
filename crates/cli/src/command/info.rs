use std::{io, path};
use std::error::Error;
use std::path::PathBuf;
use colored::Colorize;
use crate::constants;

pub async fn execute(db_path: Option<PathBuf>) -> Result<(), Box<dyn Error>> {
    let db_path = db_path.ok_or("no database found")?;
    let root_path = determine_root(&db_path)?;

    println!("Database path: {}", db_path.display().to_string().green());
    println!("Root path: {}", root_path.display().to_string().green());

    //TODO open database
    //TODO gather stats

    Ok(())
}

fn determine_root(path: &PathBuf) -> Result<PathBuf, io::Error> {
    let abs_path = path::absolute(path)?;

    if let Some(parent) = abs_path.parent() {
        if let Some(filename) = parent.file_name() {
            if filename == constants::APPLICATION_DIRECTORY {
                return Ok(parent.to_path_buf());
            }
        }
    }

    Ok(PathBuf::from(path::MAIN_SEPARATOR_STR))
}
