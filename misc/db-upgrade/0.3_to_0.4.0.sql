BEGIN TRANSACTION;

CREATE TEMPORARY TABLE file_tag_backup(file_id, tag_id);

INSERT INTO file_tag_backup
SELECT file_id, tag_id
FROM file_tag;

DROP TABLE file_tag;

CREATE TABLE file_tag (
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    value_id INTEGER NOT NULL,
    PRIMARY KEY (file_id, tag_id, value_id),
    FOREIGN KEY (file_id) REFERENCES file(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
    FOREIGN KEY (value_id) REFERENCES value(id));

INSERT INTO file_tag
SELECT file_id, tag_id, 0
FROM file_tag_backup;

COMMIT;
