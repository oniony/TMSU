use std::error::Error;
use std::fs;
use std::path::PathBuf;
use rusqlite::Connection;
use crate::migrations;

pub fn create(path: &PathBuf) -> Result<(), Box<dyn Error>> {
    if path.exists() {
        return Err(format!("{}: Database already exists", path.to_str().unwrap()).into())
    }

    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)?;
    }

    let mut conn = Connection::open(path)?;
    let mut tx = conn.transaction()?;

    migrations::run(&mut tx)?;
    tx.commit()?;

    Ok(())
}
