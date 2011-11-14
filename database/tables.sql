---------
-- tag --
---------

DROP TABLE IF EXISTS tag;

CREATE TABLE tag
(
    id INTEGER PRIMARY KEY,
    name TEXT NOT NULL
);

----------
-- file --
----------

DROP TABLE IF EXISTS file;

CREATE TABLE file
(
    id INTEGER PRIMARY KEY,
    path TEXT UNIQUE NOT NULL,
    fingerprint TEXT NOT NULL
);

--------------
-- file-tag --
--------------

DROP TABLE IF EXISTS file_tag;

CREATE TABLE file_tag
(
    id INTEGER PRIMARY KEY,
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL,
    FOREIGN KEY (file_id) REFERENCES file(id),
    FOREIGN KEY (tag_id) REFERENCES tag(id)
);
