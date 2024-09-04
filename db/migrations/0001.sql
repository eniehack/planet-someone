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

-- +migrate Down
DROP TABLE posts;