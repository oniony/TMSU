ALTER TABLE file_tag ADD COLUMN explicit NOT NULL DEFAULT 1;
ALTER TABLE file_tag ADD COLUMN implicit NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS tmsu (
    schema_version TEXT
);
INSERT INTO tmsu (schema_version) VALUES('0.1.0');
