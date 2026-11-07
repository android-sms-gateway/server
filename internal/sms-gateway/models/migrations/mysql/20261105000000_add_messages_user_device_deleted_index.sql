-- +goose Up
-- +goose StatementBegin
CREATE INDEX `idx_messages_user_device_deleted` ON `messages` (`user_id`, `device_id`, `deleted_at`);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP INDEX `idx_messages_user_device_deleted` ON `messages`;
-- +goose StatementEnd