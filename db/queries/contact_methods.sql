-- name: CreateContactMethod :one
INSERT INTO contact_methods (contact_id, type, value)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListContactMethodsByContact :many
SELECT contact_methods.*
FROM contact_methods
JOIN contacts ON contacts.id = contact_methods.contact_id
WHERE contact_methods.contact_id = $1 AND contacts.user_id = $2;

-- name: DeleteContactMethod :exec
DELETE FROM contact_methods
USING contacts
WHERE contact_methods.id = $1
  AND contact_methods.contact_id = contacts.id
  AND contacts.user_id = $2;