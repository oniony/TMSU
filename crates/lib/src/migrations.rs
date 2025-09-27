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
    let version = get_schema_version(tx)?.unwrap_or(0);

    if version < 1 {
        create_schema(tx)?;
    }

    Ok(())
}

fn create_version_table<'t>(tx: &mut Transaction<'t>) -> Result<(), Box<dyn std::error::Error>> {
    let sql = include_str!("migrations/0_schema_version.sql");
    tx.execute(sql, ())?;

    Ok(())
}

fn get_schema_version<'t>(tx: &mut Transaction<'t>) -> Result<Option<u32>, Box<dyn std::error::Error>> {
    let sql = "SELECT version FROM schema_version";

    let version = tx
        .query_row(sql, [], |row| row.get(0))
        .optional()?;

    Ok(version)
}

fn create_schema<'t>(tx: &mut Transaction<'t>) -> Result<(), Box<dyn std::error::Error>> {
    let sql = include_str!("migrations/1_schema.sql");
    tx.execute_batch(sql)?;

    Ok(())
}