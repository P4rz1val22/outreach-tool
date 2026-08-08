-- name: CreateCheckIn :one
INSERT INTO check_ins (thread_id, date, deadline)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetCurrentPendingCheckIn :one
SELECT check_ins.*
FROM check_ins
JOIN threads ON threads.id = check_ins.thread_id
JOIN contacts ON contacts.id = threads.contact_id
WHERE check_ins.thread_id = $1
  AND check_ins.status = 'pending'
  AND contacts.user_id = $2;

-- name: ResolveCheckIn :one
UPDATE check_ins
SET status = $2, resolved_at = now()
FROM threads, contacts
WHERE check_ins.id = $1
  AND check_ins.thread_id = threads.id
  AND threads.contact_id = contacts.id
  AND contacts.user_id = $3
RETURNING check_ins.*;

-- name: RescheduleCheckIn :one
UPDATE check_ins
SET date = $2, deadline = $3
FROM threads, contacts
WHERE check_ins.id = $1
  AND check_ins.thread_id = threads.id
  AND threads.contact_id = contacts.id
  AND contacts.user_id = $4
  AND check_ins.status = 'pending'
RETURNING check_ins.*;