-- +goose Up
-- +goose StatementBegin
CREATE INDEX `idx_messages_created_at` ON `messages` (`created_at`);
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX `idx_messages_device_created_at` ON `messages` (`device_id`, `created_at`);
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX `idx_messages_is_hashed` ON `messages`;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX `idx_messages_unhashed` ON `messages` (`is_hashed`, `is_encrypted`, `state`);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP INDEX `idx_messages_unhashed` ON `messages`;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX `idx_messages_is_hashed` USING HASH ON `messages` (`is_hashed`);
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX `idx_messages_device_created_at` ON `messages`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP INDEX `idx_messages_created_at` ON `messages`;
-- +goose StatementEnd
