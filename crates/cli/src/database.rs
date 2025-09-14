use std::io;
use std::io::{Error,ErrorKind};
use std::path::PathBuf;
use crate::constants;

pub fn resolve(path: Option<PathBuf>) -> Result<PathBuf, Error> {
    match path {
        Some(path) => Ok(path),
        None => match find() {
            Ok(Some(path)) => Ok(path),
            Ok(None) => Err(Error::new(ErrorKind::NotFound, "no database found: use 'tmsu init' to create one")),
            Err(err) => Err(err),
        }
    }
}

fn find() -> Result<Option<PathBuf>, io::Error> {
    let mut search = std::env::current_dir()?;

    loop {
        let candidate = search
            .join(constants::APPLICATION_DIRECTORY)
            .join(constants::DEFAULT_DATABASE_NAME);

        if candidate.exists() {
            return Ok(Some(candidate))
        }

        if !search.pop() {
            return Ok(None);
        }
    }
}
