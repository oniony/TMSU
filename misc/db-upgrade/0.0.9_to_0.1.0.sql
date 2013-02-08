-- idx_file_path is redundant and an overhead as the unique constraint creates an identical index
DROP INDEX IF EXISTS idx_file_path;

-- file_tag table is replaced by separate tables for explicit and implicit taggings

CREATE TABLE IF NOT EXISTS explicit_file_tag (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id)
    FOREIGN KEY (file_id) REFERENCES file(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
);

CREATE INDEX IF NOT EXISTS idx_explicit_file_tag_file_id
ON explicit_file_tag(file_id);

CREATE INDEX IF NOT EXISTS idx_explicit_file_tag_tag_id
ON explicit_file_tag(tag_id);

CREATE TABLE IF NOT EXISTS implicit_file_tag (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id)
    FOREIGN KEY (file_id) REFERENCES file(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
);

CREATE INDEX IF NOT EXISTS idx_implicit_file_tag_file_id
ON implicit_file_tag(file_id);

CREATE INDEX IF NOT EXISTS idx_implicit_file_tag_tag_id
ON implicit_file_tag(tag_id);

INSERT INTO implicit_file_tag SELECT file_id, tag_id
                              FROM file_tag
                              WHERE implicit;

DELETE FROM file_tag
WHERE NOT explicit;

INSERT INTO explicit_file_tag SELECT file_id, tag_id
                              FROM file_tag;

DROP TABLE file_tag;

-- new size column on the file table
ALTER TABLE file ADD COLUMN size INTEGER NOT NULL DEFAULT 0;

-- invalid dates cause a problem for go-sqlite 3 so need to set them to epoch.
UPDATE file SET mod_time = '1970-01-01 00:00:00' WHERE mod_time = '0000-00-00 00:00:00';

-- the new size field does not get repaired by the 'repair' command unless it thinks the file has changed so
-- set the mod_time to epoch for all files with a zero size to force the 'repair' command to update them.
UPDATE file SET mod_time = '1970-01-01 00:00:00' WHERE size = 0;
