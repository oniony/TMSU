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

CREATE TABLE IF NOT EXISTS tag (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_tag_name
ON tag(name);

-- file

CREATE TABLE IF NOT EXISTS file (
    id INTEGER PRIMARY KEY,
    directory TEXT NOT NULL,
    name TEXT NOT NULL,
    fingerprint TEXT NOT NULL,
    mod_time DATETIME NOT NULL,
    size INTEGER NOT NULL,
    is_dir BOOLEAN NOT NULL,
    CONSTRAINT con_file_path UNIQUE (directory, name)
);

CREATE INDEX IF NOT EXISTS idx_file_fingerprint
ON file(fingerprint);

-- value

CREATE TABLE IF NOT EXISTS value (
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL,
    CONSTRAINT con_value_name UNIQUE (name)
);

-- file-tag

CREATE TABLE IF NOT EXISTS file_tag (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    value_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id, value_id),
    FOREIGN KEY (file_id) REFERENCES file(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
    FOREIGN KEY (value_id) REFERENCES value(id)
);

CREATE INDEX IF NOT EXISTS idx_file_tag_file_id
ON file_tag(file_id);

CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id
ON file_tag(tag_id);

CREATE INDEX IF NOT EXISTS idx_file_tag_value_id
ON file_tag(value_id);

-- implication

CREATE TABLE IF NOT EXISTS implication (
    tag_id INTEGER NOT NULL,
    value_id INTEGER NOT NULL,
    implied_tag_id INTEGER NOT NULL,
    implied_value_id INTEGER NOT NULL,
    PRIMARY KEY (tag_id, value_id, implied_tag_id, implied_value_id)
);

-- query

CREATE TABLE IF NOT EXISTS query (
    text TEXT PRIMARY KEY
);

-- setting

CREATE TABLE IF NOT EXISTS setting (
    name TEXT PRIMARY KEY,
    value TEXT NOT NULL
);

-- schema version

UPDATE schema_version SET version = 1;