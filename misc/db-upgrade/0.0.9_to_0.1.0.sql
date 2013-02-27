-- idx_file_path is redundant and an overhead as the unique constraint creates an identical index
DROP INDEX IF EXISTS idx_file_path;

-- new size column on the file table
ALTER TABLE file ADD COLUMN size INTEGER NOT NULL DEFAULT 0;
ALTER TABLE file ADD COLUMN is_dir BOOLEAN NOT NULL DEFAULT 0;

-- invalid dates cause a problem for go-sqlite 3 so need to set them to epoch.
UPDATE file SET mod_time = '1970-01-01 00:00:00' WHERE mod_time = '0000-00-00 00:00:00';

-- the new size and is_dir fields do not get repaired by the 'repair' command unless it thinks the file has changed
UPDATE file SET mod_time = '1970-01-01 00:00:00';

-- drop the id column on file_tag in favour of compound PK (around the houses as SQLite3 doesn't support dropping of columns)
CREATE TABLE IF NOT EXISTS file_tag_temp (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id),
    FOREIGN KEY (file_id) REFERENCES file(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
);

INSERT INTO file_tag_temp SELECT file_id, tag_id FROM file_tag;
DROP TABLE file_tag;
ALTER TABLE file_tag_temp RENAME TO file_tag;
CREATE INDEX idx_file_tag_file_id ON file_tag(file_id);
CREATE INDEX idx_file_tag_tag_id ON file_tag(tag_id);
