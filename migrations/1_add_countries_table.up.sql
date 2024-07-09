CREATE TABLE countries (
    id SERIAL PRIMARY KEY,
    name TEXT,
    alpha2 TEXT,
    alpha3 TEXT,
    region TEXT
);
CREATE INDEX IF NOT EXISTS alpha2_idx ON countries(alpha2);