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

use std::error::Error;
use std::fmt::Display;

/// Container for multiple errors.
#[derive(Debug)]
pub struct MultiError {
    pub errors: Vec<Box<dyn Error + Send + Sync>>,
}

impl Iterator for MultiError {
    type Item = Box<dyn Error + Send + Sync>;

    fn next(&mut self) -> Option<Self::Item> {
        self.errors.pop()
    }
}

impl Error for MultiError {}

impl Display for MultiError {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        for error in &self.errors {
            f.write_fmt(format_args!("{}", error))?;
        }

        Ok(())
    }
}
