-- +goose Up
-- +goose StatementBegin
CREATE TABLE posts (
       id INTEGER PRIMARY KEY,
       title text DEFAULT "untitled",
       text text NOT NULL,
       public BOOLEAN NOT NULL CHECK (public IN (0,1)),
       read_id INTEGER UNIQUE,
       write_id INTEGER UNIQUE,
       reported BOOLEAN NOT NULL CHECK (public IN (0,1)) DEFAULT 0,
       report_reason text);
CREATE TABLE logs (
       id INTEGER PRIMARY KEY,
       timestamp DATETIME DEFAULT (datetime(CURRENT_TIMESTAMP, 'localtime')),
       message text);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE posts;
DROP TABLE logs;
-- +goose StatementEnd
