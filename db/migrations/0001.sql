-- +migrate Up
PRAGMA foreign_keys=true;

CREATE TABLE posts (
    id text,
    title text NOT NULL,
    url text NOT NULL, 
    src text NOT NULL,
    date integer NOT NULL,
    PRIMARY KEY (date, url)
);

CREATE TABLE crawl_time (
    timestamp INTEGER,
    source INTEGER NOT NULL PRIMARY KEY,
    FOREIGN KEY (source) REFERENCES sources(id)
);
-- +migrate Down
DROP TABLE posts;