CREATE TABLE posts (
    id BIGSERIAL PRIMARY KEY,
    author_fk INT REFERENCES users(id),
    content TEXT NOT NULL,
    images_urls TEXT[] NULL,
    published_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NULL
);