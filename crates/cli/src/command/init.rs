use std::env;
use std::error::Error;
use std::path::PathBuf;
use libtmsu::database;
use crate::constants::*;

pub fn execute(db_path: Option<PathBuf>) -> Result<(), Box<dyn Error>> {
    let db_path = match db_path {
        Some(path) => path,
        None => env::current_dir()?
            .join(APPLICATION_DIRECTORY)
            .join(DEFAULT_DATABASE_NAME),
    };

    database::create(db_path)
}
