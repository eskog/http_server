-- +goose Up
CREATE TABLE chirps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    body TEXT NOT NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE NOT NULL
);


-- +goose Down
DROP TABLE chirps;