-- +goose Up
-- +goose StatementBegin
ALTER TABLE `devices`
ADD `public_key` text NULL;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `devices`
ADD `key_version` int NULL DEFAULT NULL;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `devices` DROP `key_version`;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `devices` DROP `public_key`;
-- +goose StatementEnd
