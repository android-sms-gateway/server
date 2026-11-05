-- +goose Up
-- +goose StatementBegin
DROP INDEX `idx_messages_unhashed` ON `messages`;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
CREATE INDEX `idx_messages_unhashed` ON `messages` (`is_hashed`, `is_encrypted`, `state`);
-- +goose StatementEnd
