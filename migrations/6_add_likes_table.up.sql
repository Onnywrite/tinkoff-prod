CREATE TABLE likes (
    user_fk INT REFERENCES users(id),
    post_fk INT REFERENCES posts(id),
    liked_at TIMESTAMP NOT NULL DEFAULT NOW(),
    PRIMARY KEY (user_fk, post_fk)
);

CREATE INDEX likes_user_fk_idx ON likes(user_fk);
CREATE INDEX likes_post_fk_idx ON likes(post_fk);
CREATE INDEX likes_liked_at_idx ON likes USING btree (liked_at DESC);