-- name: CreateUser :one
INSERT INTO app_user (email, password) VALUES ($1, $2) RETURNING id;

-- name: DoesUserWithEmailExist :one
SELECT EXISTS(SELECT 1 FROM app_user WHERE email = $1);

-- name: GetUserByEmail :one
SELECT * FROM app_user WHERE email = $1;

-- name: GetUserById :one
SELECT email, created_at FROM app_user WHERE id = $1;