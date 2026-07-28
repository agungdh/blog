-- +goose Up
-- +goose StatementBegin
CREATE
VIRTUAL TABLE posts_fts USING fts5(title, markdown);
INSERT INTO posts_fts(rowid, title, markdown)
SELECT id, title, markdown
FROM posts;
CREATE TRIGGER posts_ai
    AFTER INSERT
    ON posts
BEGIN
    INSERT INTO posts_fts(rowid, title, markdown) VALUES (new.id, new.title, new.markdown);
END;
CREATE TRIGGER posts_ad
    AFTER DELETE
    ON posts
BEGIN
    DELETE FROM posts_fts WHERE rowid = old.id;
END;
CREATE TRIGGER posts_au
    AFTER UPDATE
    ON posts
BEGIN
    DELETE FROM posts_fts WHERE rowid = old.id;
    INSERT INTO posts_fts(rowid, title, markdown) VALUES (new.id, new.title, new.markdown);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS posts_au;
DROP TRIGGER IF EXISTS posts_ad;
DROP TRIGGER IF EXISTS posts_ai;
DROP TABLE IF EXISTS posts_fts;
-- +goose StatementEnd
