-- +goose Up
-- +goose StatementBegin
CREATE
VIRTUAL TABLE categories_fts USING fts5(name);
INSERT INTO categories_fts(rowid, name)
SELECT id, name
FROM categories;
CREATE TRIGGER categories_ai
    AFTER INSERT
    ON categories
BEGIN
    INSERT INTO categories_fts(rowid, name) VALUES (new.id, new.name);
END;
CREATE TRIGGER categories_ad
    AFTER DELETE
    ON categories
BEGIN
    DELETE FROM categories_fts WHERE rowid = old.id;
END;
CREATE TRIGGER categories_au
    AFTER UPDATE
    ON categories
BEGIN
    DELETE FROM categories_fts WHERE rowid = old.id;
    INSERT INTO categories_fts(rowid, name) VALUES (new.id, new.name);
END;

CREATE
VIRTUAL TABLE tags_fts USING fts5(name);
INSERT INTO tags_fts(rowid, name)
SELECT id, name
FROM tags;
CREATE TRIGGER tags_ai
    AFTER INSERT
    ON tags
BEGIN
    INSERT INTO tags_fts(rowid, name) VALUES (new.id, new.name);
END;
CREATE TRIGGER tags_ad
    AFTER DELETE
    ON tags
BEGIN
    DELETE FROM tags_fts WHERE rowid = old.id;
END;
CREATE TRIGGER tags_au
    AFTER UPDATE
    ON tags
BEGIN
    DELETE FROM tags_fts WHERE rowid = old.id;
    INSERT INTO tags_fts(rowid, name) VALUES (new.id, new.name);
END;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TRIGGER IF EXISTS tags_au;
DROP TRIGGER IF EXISTS tags_ad;
DROP TRIGGER IF EXISTS tags_ai;
DROP TABLE IF EXISTS tags_fts;
DROP TRIGGER IF EXISTS categories_au;
DROP TRIGGER IF EXISTS categories_ad;
DROP TRIGGER IF EXISTS categories_ai;
DROP TABLE IF EXISTS categories_fts;
-- +goose StatementEnd
