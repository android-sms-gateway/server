-- +goose Up
-- +goose StatementBegin
ALTER TABLE `messages`
ADD COLUMN `user_id` varchar(32) NULL
AFTER `device_id`;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `messages` `m`
    JOIN `devices` `d` ON `m`.`device_id` = `d`.`id`
SET `m`.`user_id` = `d`.`user_id`;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `messages`
MODIFY `user_id` varchar(32) NOT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX `idx_messages_user_created_at` ON `messages` (`user_id`, `created_at`, `id`);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP INDEX `idx_messages_user_created_at` ON `messages`;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `messages` DROP COLUMN `user_id`;
-- +goose StatementEnd