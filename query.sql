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

-- name: ListDocuments :many
SELECT * FROM documents
ORDER BY created_at DESC;

-- name: UpdateDocument :one
UPDATE documents
SET title = $2, content = $3, updated_at = NOW()
WHERE id = $1 AND author_id = $4
RETURNING *;

-- name: DeleteDocument :exec
DELETE FROM documents 
WHERE id = $1 AND author_id = $2;

-- name: AddTeamMember :one
INSERT INTO team_members (user_id, team_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListTeamMembers :many
SELECT users.id, users.name, team_members.role, team_members.joined_at
FROM users
JOIN team_members ON users.id = team_members.user_id
WHERE team_members.team_id = $1
ORDER BY team_members.joined_at DESC;