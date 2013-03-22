-- add the new implication table
CREATE TABLE IF NOT EXISTS implication (
    tag_id INTEGER NOT NULL,
    implied_tag_id INTEGER_NOT_NULL,
    PRIMARY KEY (tag_id, implied_tag_id)
);
