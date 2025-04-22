-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, email, hashed_password)
VALUES (
	$1,
	$2,
	$3,
	$4,
	$5
)
RETURNING *;

-- name: DeleteUsers :exec
DELETE FROM users;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = $1
LIMIT 1;

-- name: GetUserFromRefreshToken :one
SELECT * FROM users
INNER JOIN refresh_tokens
ON users.id = refresh_tokens.user_id
WHERE refresh_tokens.token = $1
LIMIT 1;

-- name: UpdateUserEmailAndPassword :one
UPDATE users
SET email = $1, hashed_password = $2
WHERE id = $3
RETURNING *;

-- name: UpgradeUserToRedByID :exec
UPDATE users
SET is_chirpy_red = true
WHERE id = $1
RETURNING *;
