-- new columns
ALTER TABLE file ADD COLUMN size INTEGER NOT NULL DEFAULT 0;
ALTER TABLE file_tag ADD COLUMN explicit BOOL NOT NULL DEFAULT 1;
ALTER TABLE file_tag ADD COLUMN implicit BOOL NOT NULL DEFAULT 0;

-- invalid dates cause a problem for go-sqlite 3 so need to set them to epoch.
UPDATE file SET mod_time = '1970-01-01 00:00:00' WHERE mod_time = '0000-00-00 00:00:00' or size = 0;

-- the new size field does not get repaired by the 'repair' command unless it thinks the file has changed so
-- set the mod_time to epoch for all files with a zero size to force the 'repair' command to update them.
UPDATE file SET mod_time = '1970-01-01 00:00:00' WHERE size = 0;
