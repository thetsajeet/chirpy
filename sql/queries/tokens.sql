-- name: StoreRefreshToken :exec
insert into refresh_tokens (token, created_at, updated_at, user_id, expires_at, revoked_at)
values ($1, now(), now(), $2, $3, null);

-- name: LookupToken :one
select *
from refresh_tokens
where token = $1;

-- name: RevokeToken :exec
update refresh_tokens
set updated_at = now(), revoked_at = now()
where token = $1;