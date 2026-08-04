-- name: ListContacts :many
SELECT * FROM contacts
WHERE status = 'active' AND user_id = $1
ORDER BY name;

-- name: ListContactsByName :many
SELECT *
FROM contacts
WHERE name = $1 AND user_id = $2;

-- name: GetContactByID :one
SELECT *
FROM contacts
WHERE id = $1 AND user_id = $2;

-- name: CreateContact :one
INSERT INTO contacts (name, role, user_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateContact :one
UPDATE contacts
SET name = $1, role = $2, updated_at = now()
WHERE id = $3 AND user_id = $4
RETURNING *;

-- name: ArchiveContact :exec
UPDATE contacts
SET status = 'archived', updated_at = now()
WHERE id = $1 AND user_id = $2;