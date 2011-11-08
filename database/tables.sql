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
    fingerprint TEXT UNIQUE NOT NULL
);

---------------
-- file-path --
---------------

DROP TABLE IF EXISTS file_path;

CREATE TABLE file_path
(
    id INTEGER PRIMARY KEY,
    file_id INTEGER NOT NULL,
    path TEXT UNIQUE NOT NULL
);

--------------
-- file-tag --
--------------

DROP TABLE IF EXISTS file_tag;

CREATE TABLE file_tag
(
    id INTEGER PRIMARY KEY,
    file_id INTEGER NOT NULL,
    tag_id INTEGER NOT NULL
);
