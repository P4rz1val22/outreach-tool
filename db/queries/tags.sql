-- name: CreateTag :one
INSERT INTO tags (name, user_id)
VALUES ($1, $2)
RETURNING *;

-- name: ListTags :many
SELECT * FROM tags
WHERE user_id = $1
ORDER BY name;

-- name: AddTagToContact :exec
INSERT INTO contact_tags (contact_id, tag_id)
SELECT $1, $2
WHERE EXISTS (SELECT 1 FROM contacts WHERE contacts.id = $1 AND contacts.user_id = $3)
  AND EXISTS (SELECT 1 FROM tags WHERE tags.id = $2 AND tags.user_id = $3);

-- name: RemoveTagFromContact :exec
DELETE FROM contact_tags
USING contacts
WHERE contact_tags.contact_id = $1
  AND contact_tags.tag_id = $2
  AND contact_tags.contact_id = contacts.id
  AND contacts.user_id = $3;

-- name: ListTagsForContact :many
SELECT tags.*
FROM tags
JOIN contact_tags ON contact_tags.tag_id = tags.id
JOIN contacts ON contacts.id = contact_tags.contact_id
WHERE contact_tags.contact_id = $1
  AND tags.user_id = $2
  AND contacts.user_id = $2;

-- name: AddTagToThread :exec
INSERT INTO thread_tags (thread_id, tag_id)
SELECT $1, $2
WHERE EXISTS (
    SELECT 1 FROM threads JOIN contacts ON contacts.id = threads.contact_id
    WHERE threads.id = $1 AND contacts.user_id = $3
  )
  AND EXISTS (SELECT 1 FROM tags WHERE tags.id = $2 AND tags.user_id = $3);

-- name: RemoveTagFromThread :exec
DELETE FROM thread_tags
USING threads, contacts
WHERE thread_tags.thread_id = $1
  AND thread_tags.tag_id = $2
  AND thread_tags.thread_id = threads.id
  AND threads.contact_id = contacts.id
  AND contacts.user_id = $3;

-- name: ListTagsForThread :many
SELECT tags.*
FROM tags
JOIN thread_tags ON thread_tags.tag_id = tags.id
JOIN threads ON threads.id = thread_tags.thread_id
JOIN contacts ON contacts.id = threads.contact_id
WHERE thread_tags.thread_id = $1
  AND tags.user_id = $2
  AND contacts.user_id = $2;