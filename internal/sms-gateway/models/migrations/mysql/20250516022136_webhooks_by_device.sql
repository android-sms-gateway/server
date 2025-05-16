-- +goose Up
-- +goose StatementBegin
ALTER TABLE `webhooks`
ADD `device_id` char(21);
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `webhooks`
ADD CONSTRAINT `fk_webhooks_device` FOREIGN KEY (`device_id`) REFERENCES `devices`(`id`) ON DELETE CASCADE;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX `idx_webhooks_device` ON `webhooks`(`device_id`);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `webhooks` DROP FOREIGN KEY `fk_webhooks_device`;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `webhooks` DROP `device_id`;
-- +goose StatementEnd