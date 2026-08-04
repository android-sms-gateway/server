-- +goose Up
-- +goose StatementBegin
ALTER TABLE `message_recipients`
MODIFY COLUMN `phone_number` varchar(512) NOT NULL;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `message_recipients`
MODIFY COLUMN `phone_number` varchar(128) NOT NULL;
-- +goose StatementEnd