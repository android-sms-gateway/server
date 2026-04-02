-- +goose Up
-- +goose StatementBegin
UPDATE `devices`
SET `push_token` = NULL
WHERE `push_token` = '';
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd