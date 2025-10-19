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
use std::path::PathBuf;

/// Files command executor.
pub struct FilesCommand {
    database: Database,
    separator: Separator,
    verbosity: u8,
    args: Vec<String>,
    count: bool,
    directory: bool,
    explicit: bool,
    file: bool,
    ignore_case: bool,
    path: Option<PathBuf>,
}

impl FilesCommand {
    /// Creates a new FilesCommand.
    pub fn new(
        database: Database,
        separator: Separator,
        verbosity: u8,
        args: Vec<String>,
        count: bool,
        directory: bool,
        explicit: bool,
        file: bool,
        ignore_case: bool,
        path: Option<PathBuf>,
    ) -> FilesCommand {
        FilesCommand {
            database,
            separator,
            verbosity,
            args,
            count,
            file,
            directory,
            explicit,
            ignore_case,
            path,
        }
    }

    fn query(&self) -> String {
        self.args.join(" ")
    }

    fn casing(&self) -> Casing {
        if self.ignore_case {
            Casing::Insensitive
        } else {
            Casing::Sensitive
        }
    }

    fn tag_specificity(&self) -> TagSpecificity {
        if self.explicit {
            TagSpecificity::ExplicitOnly
        } else {
            TagSpecificity::All
        }
    }

    fn file_type(&self) -> FileTypeSpecificity {
        if self.file && !self.directory {
            FileTypeSpecificity::FileOnly
        } else if self.directory && !self.file {
            FileTypeSpecificity::DirectoryOnly
        } else {
            FileTypeSpecificity::Any
        }
    }

    /// Shows the count of files matching the expression.
    fn show_count(&self, query: &str) -> Result<(), Box<dyn Error>> {
        let count = self.database.files().query_count(
            query,
            &self.tag_specificity(),
            &self.file_type(),
            &self.casing(),
            self.path.as_ref(),
        )?;

        print!("{}{}", count, self.separator);

        Ok(())
    }

    /// Shows the files matching the expression.
    fn show_files(&self, query: &str) -> Result<(), Box<dyn Error>> {
        let files = self.database.files().query(
            query,
            &self.tag_specificity(),
            &self.file_type(),
            &self.casing(),
            self.path.as_ref(),
        )?;

        for file in files {
            print!("{}{}", file.path().to_str().unwrap_or(""), self.separator);
        }

        Ok(())
    }
}

impl Executor for FilesCommand {
    fn execute(&self) -> Result<(), Box<dyn Error>> {
        let query = self.query();

        if self.verbosity > 0 {
            println!("query: {}", query);
        }

        if self.count {
            self.show_count(&query)
        } else {
            self.show_files(&query)
        }
    }
}
