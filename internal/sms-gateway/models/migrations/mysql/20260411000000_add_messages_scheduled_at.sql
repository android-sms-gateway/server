-- +goose Up
-- +goose StatementBegin
ALTER TABLE `messages`
ADD `schedule_at` datetime NULL DEFAULT NULL;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `messages` DROP `schedule_at`;
-- +goose StatementEnd