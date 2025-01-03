-- Create the extension and indexes for full-text search exercise 36
-- Check article: https://niallburkley.com/blog/index-columns-for-like-in-postgres/
CREATE EXTENSION IF NOT EXISTS pg_trgm;

CREATE INDEX idx_comments_content ON comments USING gin (content gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_title ON posts USING gin (title gin_trgm_ops);
CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin (tags);

CREATE INDEX IF NOT EXISTS idx_users_username ON users (username);
CREATE INDEX IF NOT EXISTS idx_posts_user_id ON posts (user_id);
CREATE INDEX IF NOT EXISTS idx_comments_post_id ON comments (post_id);

/*
we need gin on tags for CREATE INDEX IF NOT EXISTS idx_posts_tags ON posts USING gin (tags); as 
not only we do full text search on texts for content in first command but also on tags which is an array
When we start working on user feed this module ex 36 and next module 8 we rely doing search on filtering so important step 

we have added indexes for username, userid and postid without gin as we have methods for retrieving posts, users and usernames
by its ID with respective table name
*/