CREATE INDEX author_fk_idx ON posts(author_fk);
CREATE INDEX published_at_idx ON posts USING btree (published_at DESC);