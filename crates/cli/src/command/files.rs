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
use libtmsu::common::Casing;
use libtmsu::database::Database;
use libtmsu::file::FileTypeSpecificity;
use libtmsu::tag::TagSpecificity;
use std::error::Error;

/// Files command executor.
pub struct FilesCommand {
    database: Database,
    args: Vec<String>,
    separator: Separator,
    count: bool,
    tag_specificity: TagSpecificity,
    file_type: FileTypeSpecificity,
    casing: Casing,
}

impl FilesCommand {
    /// Creates a new FilesCommand.
    pub fn new(
        database: Database,
        args: Vec<String>,
        separator: Separator,
        count: bool,
        tag_specificity: TagSpecificity,
        file_type: FileTypeSpecificity,
        casing: Casing,
    ) -> FilesCommand {
        FilesCommand {
            database,
            args,
            separator,
            count,
            tag_specificity,
            file_type,
            casing,
        }
    }

    /// Shows the count of files matching the expression.
    fn show_count(&self, query: &str) -> Result<(), Box<dyn Error>> {
        let count = self.database.files().query_count(
            query,
            &self.tag_specificity,
            &self.file_type,
            &self.casing,
        )?;

        print!("{}{}", count, self.separator);

        Ok(())
    }

    /// Shows the files matching the expression.
    fn show_files(&self, query: &str) -> Result<(), Box<dyn Error>> {
        let files = self.database.files().query(
            query,
            &self.tag_specificity,
            &self.file_type,
            &self.casing,
        )?;

        for file in files {
            print!("{}{}", file.path().to_str().unwrap_or(""), self.separator);
        }

        Ok(())
    }
}

impl Executor for FilesCommand {
    fn execute(&self) -> Result<(), Box<dyn Error>> {
        let query = self.args.join(" ").to_owned();

        if self.count {
            self.show_count(&query)
        } else {
            self.show_files(&query)
        }
    }
}
