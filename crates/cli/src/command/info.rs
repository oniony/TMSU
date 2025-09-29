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

use colored::Colorize;
use libtmsu::database::Database;
use std::error::Error;

/// Executes the 'info' command, which provides database information.
pub fn execute(database: Database) -> Result<(), Box<dyn Error>> {
    println!(
        "Database path: {}",
        database.path().display().to_string().green()
    );
    println!(
        "Root path: {}",
        database.root().display().to_string().green()
    );

    //TODO open database
    //TODO gather stats

    Ok(())
}
