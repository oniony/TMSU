use std::fmt;

use crate::errors::*;
use crate::storage::Transaction;

pub const LATEST_SCHEMA_VERSION: SchemaVersion = SchemaVersion {
    major: 0,
    minor: 7,
    patch: 0,
    revision: 1,
};

#[derive(Debug, PartialEq, Eq, PartialOrd, Ord)]
pub struct SchemaVersion {
    major: u32,
    minor: u32,
    patch: u32,
    revision: u32,
}

impl SchemaVersion {
    pub fn from_tuple(major: u32, minor: u32, patch: u32, revision: u32) -> Self {
        Self {
            major,
            minor,
            patch,
            revision,
        }
    }
}

impl fmt::Display for SchemaVersion {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        write!(
            f,
            "{}.{}.{}-{}",
            self.major, self.minor, self.patch, self.revision
        )
    }
}

pub fn current_schema_version(tx: &mut Transaction) -> Result<Option<SchemaVersion>> {
    let sql = "
SELECT *
FROM version";

    fn convert_row(row: super::Row) -> Result<SchemaVersion> {
        let revision = if row.column_count() > 3 {
            row.get(3)?
        } else {
            // Might be prior to schema 0.7.0-1 where there was no revision column
            0
        };
        Ok(SchemaVersion {
            major: row.get(0)?,
            minor: row.get(1)?,
            patch: row.get(2)?,
            revision,
        })
    }

    let res = tx.query_single(sql, convert_row);

    // Ignore errors
    // TODO: this is similar to the Go implementation, but it would be better to ignore only the
    // case where the table does not exist.
    Ok(res.ok().flatten())
}

pub fn insert_schema_version(tx: &mut Transaction, version: &SchemaVersion) -> Result<()> {
    let sql = "
INSERT INTO version (major, minor, patch, revision)
VALUES (?, ?, ?, ?)";

    let params = rusqlite::params![
        version.major,
        version.minor,
        version.patch,
        version.revision
    ];
    match tx.execute_params(sql, params) {
        // TODO: is all this error handling really necessary?
        Ok(1) => Ok(()),
        Ok(_) => {
            Err("Version could not be inserted: expected exactly one row to be affected".into())
        }
        Err(e) => Err(e.chain_err(|| "Could not insert schema version")),
    }
}

pub fn update_schema_version(tx: &mut Transaction, version: &SchemaVersion) -> Result<()> {
    let sql = "
UPDATE version SET major = ?, minor = ?, patch = ?, revision = ?";

    let params = rusqlite::params![
        version.major,
        version.minor,
        version.patch,
        version.revision
    ];
    match tx.execute_params(sql, params) {
        // TODO: is all this error handling really necessary?
        Ok(1) => Ok(()),
        Ok(_) => {
            Err("Version could not be updated: expected exactly one row to be affected".into())
        }
        Err(e) => Err(e.chain_err(|| "Could not update schema version")),
    }
}

pub fn create_schema(tx: &mut Transaction) -> Result<()> {
    create_tag_table(tx)?;
    create_value_table(tx)?;
    create_file_table(tx)?;
    create_file_tag_table(tx)?;
    create_implication_table(tx)?;
    create_query_table(tx)?;
    create_setting_table(tx)?;
    create_version_table(tx)?;
    insert_schema_version(tx, &LATEST_SCHEMA_VERSION)?;
    Ok(())
}

fn create_tag_table(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
CREATE TABLE IF NOT EXISTS tag (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
)",
    )?;

    tx.execute(
        "
CREATE INDEX IF NOT EXISTS idx_tag_name
ON tag(name)",
    )?;

    Ok(())
}

fn create_value_table(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
CREATE TABLE IF NOT EXISTS value (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT con_value_name UNIQUE (name)
)",
    )?;

    Ok(())
}

fn create_file_table(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
CREATE TABLE IF NOT EXISTS file (
    id INTEGER PRIMARY KEY,
    directory TEXT NOT NULL,
    name TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    mod_time DATETIME NOT NULL,
    size INTEGER NOT NULL,
    is_dir BOOLEAN NOT NULL,
    CONSTRAINT con_file_path UNIQUE (directory, name)
)",
    )?;

    tx.execute(
        "
CREATE INDEX IF NOT EXISTS idx_file_fingerprint
ON file(fingerprint)",
    )?;

    Ok(())
}

fn create_file_tag_table(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
CREATE TABLE IF NOT EXISTS file_tag (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    value_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id, value_id),
    FOREIGN KEY (file_id) REFERENCES file(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
    FOREIGN KEY (value_id) REFERENCES value(id)
)",
    )?;

    tx.execute(
        "
CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
ON file_tag(file_id)",
    )?;

    tx.execute(
        "
CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
ON file_tag(tag_id)",
    )?;

    tx.execute(
        "
CREATE INDEX IF NOT EXISTS idx_file_tag_value_id
ON file_tag(value_id)",
    )?;

    Ok(())
}

pub fn create_implication_table(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
CREATE TABLE IF NOT EXISTS implication (
    tag_id INTEGER NOT NULL,
    value_id INTEGER NOT NULL,
    implied_tag_id INTEGER NOT NULL,
    implied_value_id INTEGER NOT NULL,
    PRIMARY KEY (tag_id, value_id, implied_tag_id, implied_value_id)
)",
    )?;

    Ok(())
}

fn create_query_table(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
CREATE TABLE IF NOT EXISTS query (
    text TEXT PRIMARY KEY
)",
    )?;

    Ok(())
}

fn create_setting_table(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
CREATE TABLE IF NOT EXISTS setting (
    name TEXT PRIMARY KEY,
    value TEXT NOT NULL
)",
    )?;

    Ok(())
}

pub fn create_version_table(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
CREATE TABLE IF NOT EXISTS version (
    major NUMBER NOT NULL,
    minor NUMBER NOT NULL,
    patch NUMBER NOT NULL,
    revision NUMBER NOT NULL,
    PRIMARY KEY (major, minor, patch, revision)
)",
    )?;

    Ok(())
}
