-- name: GetUserById :one
SELECT id, username, email, role, created_at, updated_at
FROM users
WHERE id = $1;

-- name: GetUserByEmail :one
SELECT id, username, email, role, created_at, updated_at
FROM users
WHERE email = $1;

-- name: GetUserByEmailAndPassword :one
SELECT * FROM users WHERE email = $1;

-- name: CreateUser :one
INSERT INTO users(username, email, password)
VALUES ($1, $2, $3)
RETURNING id, username, email, role, created_at, updated_at;
