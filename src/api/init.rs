use std::fs;
use std::path::{Path, PathBuf};

use crate::errors::*;
use crate::storage::Storage;

pub fn run_init(path: &Path) -> Result<()> {
    println!("{}: creating database", path.display());

    // Create directory if needed
    let mut db_path = PathBuf::from(path);
    db_path.push(".tmsu");
    fs::create_dir_all(&db_path)?;

    db_path.push("db");

    Storage::create_at(&db_path)
}
