-- +migrate Up
PRAGMA foreign_keys=true;

CREATE TABLE posts (
    id text NOt NULL UNIQUE,
    content text NOT NULL,
    url text NOT NULL, 
    src text NOT NULL,
    type text NOT NULL,
    created_at integer NOT NULL,
    PRIMARY KEY (created_at, url)
);

-- +migrate Down
DROP TABLE posts;