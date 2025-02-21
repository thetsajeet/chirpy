// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.28.0
// source: users.sql

package database

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const createUser = `-- name: CreateUser :one
insert into users (id, created_at, updated_at, email, hashed_password)
values (
    gen_random_uuid(), now(), now(), $1, $2
)
returning id, created_at, updated_at, email
`

type CreateUserParams struct {
	Email          string
	HashedPassword string
}

type CreateUserRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (CreateUserRow, error) {
	row := q.db.QueryRowContext(ctx, createUser, arg.Email, arg.HashedPassword)
	var i CreateUserRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
	)
	return i, err
}

const deleteAllUsers = `-- name: DeleteAllUsers :exec
delete from users
`

func (q *Queries) DeleteAllUsers(ctx context.Context) error {
	_, err := q.db.ExecContext(ctx, deleteAllUsers)
	return err
}

const loginUser = `-- name: LoginUser :one
select id, created_at, updated_at, email, hashed_password
from users
where email = $1
`

func (q *Queries) LoginUser(ctx context.Context, email string) (User, error) {
	row := q.db.QueryRowContext(ctx, loginUser, email)
	var i User
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
		&i.HashedPassword,
	)
	return i, err
}

const updateUser = `-- name: UpdateUser :one
update users
set email = $1, hashed_password = $2, updated_at = now()
where id = $3
returning id, created_at, updated_at, email
`

type UpdateUserParams struct {
	Email          string
	HashedPassword string
	ID             uuid.UUID
}

type UpdateUserRow struct {
	ID        uuid.UUID
	CreatedAt time.Time
	UpdatedAt time.Time
	Email     string
}

func (q *Queries) UpdateUser(ctx context.Context, arg UpdateUserParams) (UpdateUserRow, error) {
	row := q.db.QueryRowContext(ctx, updateUser, arg.Email, arg.HashedPassword, arg.ID)
	var i UpdateUserRow
	err := row.Scan(
		&i.ID,
		&i.CreatedAt,
		&i.UpdatedAt,
		&i.Email,
	)
	return i, err
}
