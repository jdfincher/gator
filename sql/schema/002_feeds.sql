-- +goose Up
CREATE TABLE feeds(
  id UUID PRIMARY KEY,
  created_at TIMESTAMP NOT NULL,
  updated_at TIMESTAMP NOT NULL,
  name TEXT NOT NULL,
  url TEXT NOT NULL,
  user_id UUID NOT NULL, 
  CONSTRAINT fk_users FOREIGN KEY(user_id) REFERENCES users(id) ON DELETE CASCADE,
  CONSTRAINT feeds_user_url UNIQUE (user_id, url)
);

-- +goose Down
DROP TABLE feeds;
