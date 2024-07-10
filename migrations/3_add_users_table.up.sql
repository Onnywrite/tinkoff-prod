CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    name VARCHAR(32) NOT NULL,
    lastname VARCHAR(32) NOT NULL,
    email CHARACTER VARYING(64) UNIQUE NOT NULL,
    country_fk BIGINT REFERENCES countries(id),
    is_public BOOLEAN DEFAULT true,
    image CHARACTER VARYING(100),
    password CHAR(60) NOT NULL,
    birthday DATE NOT NULL
);
CREATE INDEX IF NOT EXISTS email_idx ON users(email);