-- name: GetUser :one
SELECT * FROM accounts
WHERE id = $1 LIMIT 1;

-- name: GetUsers :many
SELECT * FROM accounts
ORDER BY username;

-- name: GetUserByUsername :one
SELECT * FROM accounts
WHERE username = $1 LIMIT 1;

-- name: GetUserByEmail :one
SELECT * FROM accounts
WHERE email = $1 LIMIT 1;

-- name: GetTx :one
SELECT * FROM transactions
WHERE id = $1 LIMIT 1;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;

-- name: CreateAccount :one
INSERT INTO accounts (
	username, balance, email
) VALUES (
	$1, $2, $3
)
RETURNING *;

-- name: CreateTransaction :one
INSERT INTO transactions (
	from_account, to_account, amount
) VALUES (
	$1, $2, $3
)
RETURNING *;