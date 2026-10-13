CREATE TABLE IF NOT EXISTS items (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL
);

INSERT INTO items (name)
SELECT name
FROM (VALUES ('first item'), ('second item')) AS seed(name)
WHERE NOT EXISTS (SELECT 1 FROM items);
