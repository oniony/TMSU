INSERT INTO tag (id, name) VALUES (1, 'inside');
INSERT INTO tag (id, name) VALUES (2, 'outside');

INSERT INTO file (id, fingerprint) VALUES (1, '1a2b3c4d');
INSERT INTO file (id, fingerprint) VALUES (2, '5e6f7a8b');

INSERT INTO file_path (id, file_id, path) VALUES (1, 1, '/tmp/file1a');
INSERT INTO file_path (id, file_id, path) VALUES (2, 1, '/tmp/file1b');
INSERT INTO file_path (id, file_id, path) VALUES (3, 2, '/tmp/file2');

INSERT INTO file_tag (id, file_id, tag_id) VALUES (1, 1, 1);
INSERT INTO file_tag (id, file_id, tag_id) VALUES (2, 2, 2);
