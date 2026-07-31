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