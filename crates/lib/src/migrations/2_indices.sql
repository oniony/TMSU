-- Copyright 2011-2025 Paul Ruane.

-- This program is free software: you can redistribute it and/or modify
-- it under the terms of the GNU General Public License as published by
-- the Free Software Foundation, either version 3 of the License, or
-- (at your option) any later version.

-- This program is distributed in the hope that it will be useful,
-- but WITHOUT ANY WARRANTY; without even the implied warranty of
-- MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
-- GNU General Public License for more details.

-- You should have received a copy of the GNU General Public License
-- along with this program.  If not, see <http://www.gnu.org/licenses/>.

-- tag

CREATE INDEX IF NOT EXISTS idx_tag_name
ON tag(name);

-- file

CREATE INDEX IF NOT EXISTS idx_file_fingerprint
ON file(fingerprint);

-- value

-- file-tag

CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
ON file_tag(file_id);

CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
ON file_tag(tag_id);

CREATE INDEX IF NOT EXISTS idx_file_tag_value_id
ON file_tag(value_id);

-- implication

-- query

-- setting
