-- name: GenerateInvite :exec
INSERT INTO invites(user_id, invite_id, expiry) VALUES ($1, $2, $3);

-- name: DeleteInvite :exec
DELETE FROM invites WHERE user_id = $1;
