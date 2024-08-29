-- +migrate Up
PRAGMA foreign_keys=true;

CREATE TABLE sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    id_alias text NOT NULL,
    site_url text NOT NULL,
    source_url text NOT NULL,
    type int NOT NULL,
    name text NOT NULL
);

CREATE TABLE posts (
    id text PRIMARY KEY,
    title text NOT NULL,
    url text NOT NULL, 
    posts_source int NOT NULL,
    date timestamp NOT NULL,
    FOREIGN KEY (posts_source) REFERENCES sources(id)
);

CREATE TABLE crawl_time (
    timestamp INTEGER,
    source INTEGER NOT NULL,
    FOREIGN KEY (source) REFERENCES sources(id),
    PRIMARY KEY (timestamp, source)
);
-- +migrate Down
DROP TABLE posts;
DROP TABLE sources;