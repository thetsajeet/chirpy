-- +goose Up
CREATE TABLE users (
    id uuid primary key,
    created_at timestamp not null,
    updated_at timestamp not null,
    email text not null    
);

-- +goose Down
DROP TABLE users