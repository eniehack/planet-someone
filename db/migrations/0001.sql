-- +migrate Up
PRAGMA foreign_keys=true;

CREATE TABLE sources (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    site_url text NOT NULL,
    source_url text NOT NULL,
    icon_url text,
    type int NOT NULL,
    name text NOT NULL
);

CREATE UNIQUE INDEX idx__sources_id ON sources(id);

CREATE TABLE posts (
    id text,
    title text NOT NULL,
    url text NOT NULL, 
    src int NOT NULL,
    date timestamp NOT NULL,
    FOREIGN KEY (src) REFERENCES sources(id),
    PRIMARY KEY (date, url)
);

CREATE TABLE crawl_time (
    timestamp INTEGER,
    source INTEGER NOT NULL PRIMARY KEY,
    FOREIGN KEY (source) REFERENCES sources(id)
);
-- +migrate Down
DROP TABLE posts;
DROP TABLE sources;