CREATE TABLE IF NOT EXISTS goals(
  id VARCHAR(500) PRIMARY KEY,
  reddit_fullname VARCHAR(500) UNIQUE,
  reddit_link_url VARCHAR(500) NOT NULL,
  reddit_post_title VARCHAR(500) NOT NULL,
  reddit_post_created_at TIMESTAMP NOT NULL,
  s3_object_key VARCHAR(500),
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);