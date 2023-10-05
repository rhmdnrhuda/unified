CREATE TABLE IF NOT EXISTS home_module(
    id serial PRIMARY KEY,
    created_at integer,
    updated_at integer,
    title VARCHAR(255),
    subtitle VARCHAR(255),
    Banner json,
    Content json,
    Category json
);

