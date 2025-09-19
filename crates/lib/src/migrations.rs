use rusqlite::{OptionalExtension, Transaction};

pub fn run<'t>(tx: &mut Transaction<'t>) -> Result<(), Box<dyn std::error::Error>> {
    create_version_table(tx)?;
    let version = get_schema_version(tx)?.unwrap_or(0);

    if version < 1 {
        create_schema(tx)?;
    }

    Ok(())
}

pub fn create_version_table<'t>(tx: &mut Transaction<'t>) -> Result<(), Box<dyn std::error::Error>> {
    let sql = include_str!("migrations/0_schema_version.sql");
    tx.execute(sql, ())?;

    Ok(())
}

pub fn get_schema_version<'t>(tx: &mut Transaction<'t>) -> Result<Option<u32>, Box<dyn std::error::Error>> {
    let sql = "SELECT version FROM schema_version";

    let version = tx
        .query_row(sql, [], |row| row.get(0))
        .optional()?;

    Ok(version)
}

pub fn create_schema<'t>(tx: &mut Transaction<'t>) -> Result<(), Box<dyn std::error::Error>> {
    let sql = include_str!("migrations/1_schema.sql");
    tx.execute_batch(sql)?;

    Ok(())
}