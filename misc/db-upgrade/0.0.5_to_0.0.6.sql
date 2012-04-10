ALTER TABLE file ADD COLUMN mod_time DATETIME;
UPDATE file SET mod_time = '0000-00-00 00:00:00' WHERE mod_time is null;
