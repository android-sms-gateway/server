-- +goose Up
-- +goose StatementBegin
CREATE INDEX `idx_devices_last_seen` ON `devices`(`last_seen`);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP INDEX `idx_devices_last_seen` ON `devices`;
-- +goose StatementEnd