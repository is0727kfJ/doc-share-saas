-- name: CreateUser :one
INSERT INTO users (cognito_sub, name, email)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetUserByCognitoSub :one
SELECT * FROM users
WHERE cognito_sub = $1 LIMIT 1;

-- name: CreateTeam :one
INSERT INTO teams (name)
VALUES ($1)
RETURNING *;

-- name: CreateDocument :one
INSERT INTO documents (team_id, author_id, title, content)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListDocumentsByTeam :many
SELECT * FROM documents
WHERE team_id = $1
ORDER BY created_at DESC;