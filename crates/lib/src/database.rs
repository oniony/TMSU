use std::error::Error;
use std::fs;
use std::path::PathBuf;
use rusqlite::Connection;
use crate::migrations;

pub async fn create(path: PathBuf) -> Result<(), Box<dyn Error>> {
    if path.exists() {
        return Err("database already exists".into())
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
