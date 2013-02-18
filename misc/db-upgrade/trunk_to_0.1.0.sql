-- idx_file_path is redundant and an overhead as the unique constraint creates an identical index
DROP INDEX IF EXISTS idx_file_path;

-- the implicit_file_tag and explicit_file_tag tables have been merged back into file_tag
CREATE TABLE IF NOT EXISTS file_tag (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id),
    FOREIGN KEY (file_id) REFERENCES file(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
);

CREATE INDEX IF NOT EXISTS idx_file_tag_file_id ON file_tag(file_id);
CREATE INDEX IF NOT EXISTS idx_file_tag_tag_id ON file_tag(tag_id);

INSERT OR IGNORE INTO file_tag SELECT file_id, tag_id FROM explicit_file_tag;
INSERT OR IGNORE INTO file_tag SELECT file_id, tag_id FROM implicit_file_tag;

DROP TABLE explicit_file_tag;
DROP TABLE implicit_file_tag;

-- new size column on the file table
ALTER TABLE file ADD COLUMN size INTEGER NOT NULL DEFAULT 0;

-- invalid dates cause a problem for go-sqlite 3 so need to set them to epoch.
UPDATE file SET mod_time = '1970-01-01 00:00:00' WHERE mod_time = '0000-00-00 00:00:00';

-- the new size field does not get repaired by the 'repair' command unless it thinks the file has changed so
-- set the mod_time to epoch for all files with a zero size to force the 'repair' command to update them.
UPDATE file SET mod_time = '1970-01-01 00:00:00' WHERE size = 0;
