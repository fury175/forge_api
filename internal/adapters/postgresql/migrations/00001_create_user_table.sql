-- +goose Up
CREATE TABLE IF NOT EXISTS users(
    id bigserial primary key,
    username text not null,
    email text unique not null,
    password bytea not null,
    role text default 'user',
    is_active boolean default false,
    created_at timestamptz default now(),
    updated_at timestamptz default now()
);

-- +goose Down
DROP TABLE IF EXISTS users;
