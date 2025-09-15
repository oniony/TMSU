use std::env;
use std::env::VarError;
use std::error::Error;
use std::path::PathBuf;
use crate::constants;

pub fn resolve(arg_db_path: Option<PathBuf>) -> Result<Option<PathBuf>, Box<dyn Error>> {
    match arg_db_path {
        Some(path) => Ok(Some(path)),
        None => match env::var(constants::DATABASE_ENV_VAR) {
            Ok(path) => Ok(Some(PathBuf::from(path))),
            Err(VarError::NotPresent) => find(),
            Err(err) => Err(err.into()),
        }
    }
}

fn find() -> Result<Option<PathBuf>, Box<dyn Error>> {
    let working_dir = env::current_dir()?;
    let mut search = working_dir.clone();

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
