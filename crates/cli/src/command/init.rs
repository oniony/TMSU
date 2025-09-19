use std::env;
use std::error::Error;
use std::path::PathBuf;
use libtmsu::database;
use crate::constants::*;

pub fn execute(db_path: Option<PathBuf>, paths: Vec<PathBuf>) -> Result<(), Box<dyn Error>> {
    let paths = if paths.len() > 0 {
        paths
    } else if let Some(path) = db_path {
        vec![path]
    } else {
        vec![env::current_dir()?
            .join(APPLICATION_DIRECTORY)
            .join(DEFAULT_DATABASE_NAME)]
    };

    for path in paths {
        database::create(&path)?
    }

    Ok(())
}
