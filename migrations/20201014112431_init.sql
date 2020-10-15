-- +goose Up
-- +goose StatementBegin
CREATE TABLE users (
       id INTEGER PRIMARY KEY,
       username text NOT NULL UNIQUE);
CREATE TABLE posts (
       id INTEGER PRIMARY KEY,
       creator text NOT NULL,
       title text DEFAULT "untitled",
       text text NOT NULL,
       public BOOLEAN NOT NULL CHECK (public IN (0,1)),
       read_id INTEGER UNIQUE,
       write_id INTEGER UNIQUE,
       reported BOOLEAN NOT NULL CHECK (public IN (0,1)) DEFAULT 0,
       report_reason text,
       FOREIGN KEY (creator) REFERENCES users(username));
CREATE TABLE logs (
       id INTEGER PRIMARY KEY,
       timestamp DATETIME DEFAULT (datetime(CURRENT_TIMESTAMP, 'localtime')),
       message text);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE posts;
DROP TABLE users;
DROP TABLE logs;
-- +goose StatementEnd
