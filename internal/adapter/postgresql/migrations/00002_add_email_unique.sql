-- +goose Up
-- +goose StatementBegin
ALTER TABLE users ADD CONSTRAINT users_email_unique UNIQUE (email);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE users DROP CONSTRAINT IF EXISTS users_email_unique;
-- +goose StatementEnd
