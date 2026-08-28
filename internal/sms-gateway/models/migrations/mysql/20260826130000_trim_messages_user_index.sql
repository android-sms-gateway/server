-- +goose Up
-- +goose StatementBegin
DROP INDEX `idx_messages_user_created_at` ON `messages`;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX `idx_messages_user_created_at` ON `messages` (`user_id`, `created_at`);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP INDEX `idx_messages_user_created_at` ON `messages`;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE INDEX `idx_messages_user_created_at` ON `messages` (`user_id`, `created_at`, `id`);
-- +goose StatementEnd