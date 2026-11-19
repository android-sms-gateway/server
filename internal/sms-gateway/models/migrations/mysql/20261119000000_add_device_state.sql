-- +goose Up
-- +goose StatementBegin
ALTER TABLE `devices`
ADD COLUMN `state` ENUM('Active', 'Disabled') NOT NULL DEFAULT 'Active';
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `devices` DROP COLUMN `state`;
-- +goose StatementEnd