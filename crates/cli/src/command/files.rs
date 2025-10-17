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

use crate::error::MultiError;
use crate::rendering::Separator;
use crate::Executor;
use libtmsu::common::{Casing, FileTypeSpecificity, TagSpecificity};
use libtmsu::database::Database;
use libtmsu::query;
use libtmsu::query::Expression;
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
    fn show_count(&self, expression: Option<Expression>) -> Result<(), Box<dyn Error>> {
        let count = if let Some(expression) = &expression {
            self.validate_expression(&expression)?;
            self.database.files().query_count(
                expression,
                &self.tag_specificity,
                &self.file_type,
                &self.casing,
            )
        } else {
            self.database.files().all_count()
        }?;

        print!("{}{}", count, self.separator);

        Ok(())
    }

    /// Shows the files matching the expression.
    fn show_files(&self, expression: Option<Expression>) -> Result<(), Box<dyn Error>> {
        let files = if let Some(expression) = &expression {
            self.validate_expression(&expression)?;
            self.database.files().query(
                expression,
                &self.tag_specificity,
                &self.file_type,
                &self.casing,
            )
        } else {
            self.database.files().all()
        }?;

        for file in files {
            print!("{}{}", file.path().to_str().unwrap_or(""), self.separator);
        }

        Ok(())
    }

    fn validate_expression(&self, expression: &Expression) -> Result<(), Box<dyn Error>> {
        let mut errors: Vec<Box<dyn Error + Send + Sync>> = Vec::new();

        let tags = expression.tags();
        let invalid_tags = self.database.tags().missing(&tags, &self.casing)?;
        for invalid_tag in &invalid_tags {
            errors.push(format!("unknown tag: {invalid_tag}").into());
        }

        let values = expression.values();
        let invalid_values = self.database.values().missing(&values, &self.casing)?;
        for invalid_value in &invalid_values {
            errors.push(format!("unknown value: {invalid_value}").into());
        }

        if errors.is_empty() {
            Ok(())
        } else {
            Err(MultiError { errors }.into())
        }
    }
}

impl Executor for FilesCommand {
    fn execute(&self) -> Result<(), Box<dyn Error>> {
        let query_text = self.args.join(" ").to_owned();
        let expression = query::parse(&query_text)?;

        if self.count {
            self.show_count(expression)
        } else {
            self.show_files(expression)
        }
    }
}
