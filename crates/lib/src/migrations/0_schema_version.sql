-- records which version of the schema this database has been migrated to
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER
);

--INSERT INTO schema_version (version)
--VALUES (1)
--ON CONFLICT DO NOTHING;