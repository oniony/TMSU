// Copyright 2011-2025 Paul Ruane.

// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

use rusqlite::{OptionalExtension, Transaction};

/// Runs the database migrations.
pub fn run<'t>(tx: &mut Transaction<'t>) -> Result<(), Box<dyn std::error::Error>> {
    create_version_table(tx)?;

    let current_version = get_schema_version(tx)?.unwrap_or(0);

    let versions = vec![
        (1, include_str!("migrations/1_schema.sql")),
        (2, include_str!("migrations/2_indices.sql")),
    ];

    for (version, sql) in versions.iter().skip_while(|v| v.0 <= current_version) {
        run_migration(tx, *version, sql)?;
    }

    Ok(())
}

fn create_version_table<'t>(tx: &mut Transaction<'t>) -> Result<(), Box<dyn std::error::Error>> {
    tx.execute(include_str!("migrations/0_schema_version.sql"), ())?;

    Ok(())
}

fn get_schema_version<'t>(
    tx: &mut Transaction<'t>,
) -> Result<Option<u32>, Box<dyn std::error::Error>> {
    let version = tx
        .query_row("SELECT version FROM schema_version;", [], |row| row.get(0))
        .optional()?;

    Ok(version)
}

fn run_migration(
    tx: &mut Transaction,
    version: u32,
    sql: &str,
) -> Result<(), Box<dyn std::error::Error>> {
    tx.execute_batch(sql)?;
    update_schema_version(tx, version)?;

    Ok(())
}

fn update_schema_version<'t>(
    tx: &mut Transaction<'t>,
    version: u32,
) -> Result<(), Box<dyn std::error::Error>> {
    let rows = tx.execute("UPDATE schema_version SET version = ?;", [version])?;

    if rows == 0 {
        let _ = tx.execute(
            "INSERT INTO schema_version (version) VALUES (?);",
            [version],
        )?;
    };

    Ok(())
}

#[cfg(test)]
mod tests {
    use crate::migrations::{create_version_table, get_schema_version, run, update_schema_version};
    use rusqlite::Connection;

    #[test]
    fn new_database() {
        let mut conn = Connection::open_in_memory().unwrap();
        let mut tx = conn.transaction().unwrap();

        let result = run(&mut tx);

        assert!(result.is_ok());
        assert!(tx.table_exists(None, "file").unwrap());
        assert!(tx.table_exists(None, "file_tag").unwrap());
        assert!(tx.table_exists(None, "implication").unwrap());
        assert!(tx.table_exists(None, "query").unwrap());
        assert!(tx.table_exists(None, "schema_version").unwrap());
        assert!(tx.table_exists(None, "setting").unwrap());
        assert!(tx.table_exists(None, "tag").unwrap());
        assert!(tx.table_exists(None, "value").unwrap());

        let actual_version = get_schema_version(&mut tx).unwrap().unwrap();
        assert_eq!(2, actual_version);
    }

    #[test]
    fn schema_version() {
        let mut conn = Connection::open_in_memory().unwrap();
        let mut tx = conn.transaction().unwrap();

        create_version_table(&mut tx).unwrap();
        let actual_version = get_schema_version(&mut tx).unwrap();
        assert_eq!(None, actual_version);

        update_schema_version(&mut tx, 99).unwrap();
        let actual_version = get_schema_version(&mut tx).unwrap();
        assert_eq!(Some(99), actual_version);
    }
}
