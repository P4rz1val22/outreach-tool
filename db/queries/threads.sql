-- name: GetThreadByID :one
SELECT threads.*
FROM threads
JOIN contacts ON contacts.id = threads.contact_id
WHERE threads.id = $1 AND contacts.user_id = $2;

-- name: CreateThread :one
INSERT INTO threads (contact_id, label, cadence_interval_days)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListThreadsByContact :many
SELECT threads.*
FROM threads
JOIN contacts ON contacts.id = threads.contact_id
WHERE threads.contact_id = $1
  AND contacts.user_id = $2
  AND (sqlc.narg('status')::thread_status IS NULL OR threads.status = sqlc.narg('status'))
  AND (
    sqlc.narg('tag_id')::uuid IS NULL
    OR EXISTS (
        SELECT 1 FROM thread_tags
        WHERE thread_tags.thread_id = threads.id
        AND thread_tags.tag_id = sqlc.narg('tag_id')
    )
);

-- name: GetThreadByCheckInID :one
SELECT threads.*
FROM threads
JOIN check_ins ON check_ins.thread_id = threads.id
JOIN contacts ON contacts.id = threads.contact_id
WHERE check_ins.id = $1 AND contacts.user_id = $2;

-- name: ArchiveThread :exec
UPDATE threads
SET status = 'archived'
FROM contacts
WHERE threads.id = $1
  AND threads.contact_id = contacts.id
  AND contacts.user_id = $2;