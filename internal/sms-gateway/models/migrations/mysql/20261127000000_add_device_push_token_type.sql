-- +goose Up
-- +goose StatementBegin
ALTER TABLE `devices`
ADD `push_token_type` varchar(16) NULL DEFAULT NULL;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `devices` DROP `push_token_type`;
-- +goose StatementEnd