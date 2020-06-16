use crate::errors::*;
use crate::storage::schema::{self, SchemaVersion, LATEST_SCHEMA_VERSION};
use crate::storage::Transaction;

pub fn upgrade(tx: &mut Transaction) -> Result<()> {
    let current_version = schema::current_schema_version(tx)?;

    info!(
        "Database schema has version {:?}, latest schema version is {}",
        current_version, LATEST_SCHEMA_VERSION
    );

    if let Some(LATEST_SCHEMA_VERSION) = current_version {
        info!("Schema is up to date");
        return Ok(());
    }

    if current_version.is_none() {
        info!("Creating schema");
        schema::create_schema(tx)?;

        // Still need to run the upgrade, as per 0.5.0 the database did not store a version
    }

    // Use a fake 0.0.0-0 version as fallback for comparisons
    let current_version = current_version.unwrap_or(SchemaVersion::from_tuple(0, 0, 0, 0));

    info!("Upgrading database");

    if current_version < SchemaVersion::from_tuple(0, 5, 0, 0) {
        info!("Renaming fingerprint algorithm setting");
        rename_fingerprint_algorithm_setting(tx)?;
    }

    if current_version < SchemaVersion::from_tuple(0, 6, 0, 0) {
        info!("Recreating implication table");
        recreate_implication_table(tx)?;
    }

    if current_version < SchemaVersion::from_tuple(0, 7, 0, 0) {
        info!("Updating fingerprint algorithms");
        update_fingerprint_algorithms(tx)?;
    }

    if current_version < SchemaVersion::from_tuple(0, 7, 0, 1) {
        info!("Recreating version table");
        recreate_version_table(tx)?;
    }

    info!("Updating schema version");
    schema::update_schema_version(tx, &LATEST_SCHEMA_VERSION)?;

    Ok(())
}

fn rename_fingerprint_algorithm_setting(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
UPDATE setting
SET name = 'fileFingerprintAlgorithm'
WHERE name = 'fingerprintAlgorithm'",
    )?;

    Ok(())
}

fn recreate_implication_table(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
ALTER TABLE implication
RENAME TO implication_old",
    )?;

    schema::create_implication_table(tx)?;

    tx.execute(
        "
INSERT INTO implication
SELECT tag_id, 0, implied_tag_id, 0
FROM implication_old",
    )?;

    tx.execute("DROP TABLE implication_old")?;

    Ok(())
}

fn update_fingerprint_algorithms(tx: &mut Transaction) -> Result<()> {
    tx.execute(
        "
UPDATE setting
SET name = 'symlinkFingerprintAlgorithm',
    value = 'targetName'
WHERE name = 'fileFingerprintAlgorithm'
AND value = 'symlinkTargetName'",
    )?;

    tx.execute(
        "
UPDATE setting
SET name = 'symlinkFingerprintAlgorithm',
    value = 'targetNameNoExt'
WHERE name = 'fileFingerprintAlgorithm'
AND value = 'symlinkTargetNameNoExt'",
    )?;

    Ok(())
}

fn recreate_version_table(tx: &mut Transaction) -> Result<()> {
    tx.execute("DROP TABLE version")?;

    schema::create_version_table(tx)?;
    schema::insert_schema_version(tx, &LATEST_SCHEMA_VERSION)?;

    Ok(())
}
