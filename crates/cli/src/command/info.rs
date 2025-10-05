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

use crate::rendering::Separator;
use crate::Executor;
use colored::Colorize;
use libtmsu::database::Database;
use std::error::Error;

// Info command executor.
pub struct InfoCommand {
    database: Database,
    separator: Separator,
}

impl InfoCommand {
    /// Creates a new InfoCommand.
    pub fn new(database: Database, separator: Separator) -> InfoCommand {
        InfoCommand {
            database,
            separator,
        }
    }
}

impl Executor for InfoCommand {
    fn execute(&self) -> Result<(), Box<dyn Error>> {
        print!(
            "Database path: {}{}",
            self.database.path().display().to_string().green(),
            self.separator,
        );
        print!(
            "Root path: {}{}",
            self.database.root().display().to_string().green(),
            self.separator,
        );

        //TODO open database
        //TODO gather stats

        Ok(())
    }
}
