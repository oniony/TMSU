use std::error::Error;
use std::fs;
use std::path::PathBuf;
use rusqlite::Connection;

const SCHEMA_VERSION: (u32, u32, u32) = (1, 0, 0);

pub fn create(path: PathBuf) -> Result<(), Box<dyn Error>> {
    if path.exists() {
        return Err("database already exists".into())
    }

    if let Some(parent) = path.parent() {
        fs::create_dir_all(parent)?;
    }

    let mut conn = Connection::open(path)?;
    let tx = conn.transaction()?;

    //TODO schema upgrades
    create_schema(&tx)?;

    tx.commit()?;

    Ok(())
}

fn create_schema(tx: &rusqlite::Transaction<'_>) -> Result<(), Box<dyn Error>> {
    create_tag_table(tx)?;
    create_file_table(tx)?;
    create_value_table(tx)?;
    create_file_tag_table(tx)?;
    create_implication_table(tx)?;
    create_query_table(tx)?;
    create_setting_table(tx)?;
    create_version_table(tx)?;

    update_schema_version(tx, SCHEMA_VERSION)?;

    Ok(())
}

fn update_schema_version(tx: &rusqlite::Transaction<'_>, version: (u32, u32, u32)) -> Result<(), Box<dyn Error>> {
    tx.execute("
INSERT OR REPLACE INTO version (major, minor, patch) VALUES (?1, ?2, ?3);
", version)?;

    Ok(())
}

fn create_tag_table(tx: &rusqlite::Transaction<'_>) -> Result<(), Box<dyn Error>> {
    tx.execute("
CREATE TABLE IF NOT EXISTS tag (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);
", ())?;

    tx.execute("
CREATE INDEX IF NOT EXISTS idx_tag_name
ON tag (name);
", ())?;

    Ok(())
}

fn create_file_table(tx: &rusqlite::Transaction<'_>) -> Result<(), Box<dyn Error>> {
    tx.execute("
CREATE TABLE IF NOT EXISTS file (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    parent_id INTEGER,
    fingerprint TEXT NOT NULL,
    mod_time DATETIME NOT NULL,
    size INTEGER NOT NULL,
    FOREIGN KEY (parent_id) REFERENCES file (id),
    CONSTRAINT con_name_parent UNIQUE (name, parent_id)
);
", ())?;

    tx.execute("
CREATE INDEX IF NOT EXISTS idx_file_fingerprint
ON file (fingerprint);
", ())?;

    Ok(())
}

fn create_value_table(tx: &rusqlite::Transaction<'_>) -> Result<(), Box<dyn Error>> {
    tx.execute("
CREATE TABLE IF NOT EXISTS value (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT con_value_name UNIQUE (name)
);
", ())?;

    Ok(())
}

fn create_file_tag_table(tx: &rusqlite::Transaction<'_>) -> Result<(), Box<dyn Error>> {
    tx.execute("
CREATE TABLE IF NOT EXISTS file_tag (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    value_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id, value_id),
    FOREIGN KEY (file_id) REFERENCES file (id),
    FOREIGN KEY (tag_id) REFERENCES tag (id),
    FOREIGN KEY (value_id) REFERENCES value (id),
    CONSTRAINT con_file_tag UNIQUE (file_id, tag_id)
);
", ())?;

    tx.execute("
CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
ON file_tag (file_id);
", ())?;

    tx.execute("
CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
ON file_tag (tag_id);
", ())?;

    tx.execute("
CREATE INDEX IF NOT EXISTS idx_file_tag_value_id
ON file_tag (value_id);
", ())?;

    Ok(())
}

fn create_implication_table(tx: &rusqlite::Transaction<'_>) -> Result<(), Box<dyn Error>> {
    tx.execute("
CREATE TABLE IF NOT EXISTS implication (
    tag_id INTEGER NOT NULL,
    value_id INTEGER NOT NULL,
    implied_tag_id INTEGER NOT NULL,
    implied_value_id INTEGER NOT NULL,
    PRIMARY KEY (tag_id, value_id, implied_tag_id, implied_value_id)
);
", ())?;

    Ok(())
}

fn create_query_table(tx: &rusqlite::Transaction<'_>) -> Result<(), Box<dyn Error>> {
    tx.execute("
CREATE TABLE IF NOT EXISTS query (
    text TEXT PRIMARY KEY
);
", ())?;

    Ok(())
}

fn create_setting_table(tx: &rusqlite::Transaction<'_>) -> Result<(), Box<dyn Error>> {
    tx.execute("
CREATE TABLE IF NOT EXISTS setting (
    name TEXT PRIMARY KEY,
    value TEXT NOT NULL
);
", ())?;

    Ok(())
}

fn create_version_table(tx: &rusqlite::Transaction<'_>) -> Result<(), Box<dyn Error>> {
    tx.execute("
CREATE TABLE IF NOT EXISTS version (
    major INTEGER NOT NULL,
    minor INTEGER NOT NULL,
    patch INTEGER NOT NULL
);
", ())?;

    Ok(())
}