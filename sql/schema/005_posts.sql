-- +goose Up
CREATE TABLE posts(
    id UUID PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL,
    title VARCHAR(150) UNIQUE NOT NULL,
    url VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    published_at TIMESTAMP,
    feed_id UUID NOT NULL,
    FOREIGN KEY(feed_id) REFERENCES feeds (id) ON DELETE CASCADE
);

-- +goose Down
DROP TABLE posts;