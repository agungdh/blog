-- +goose Up
CREATE INDEX idx_posts_date_id ON posts(date, id);
CREATE INDEX idx_posts_category_id ON posts(category_id);
CREATE INDEX idx_post_tags_tag_id ON post_tags(tag_id);

-- +goose Down
DROP INDEX IF EXISTS idx_post_tags_tag_id;
DROP INDEX IF EXISTS idx_posts_category_id;
DROP INDEX IF EXISTS idx_posts_date_id;
