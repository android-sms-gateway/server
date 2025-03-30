-- +goose Up
-- +goose StatementBegin
ALTER TABLE `messages`
ADD `priority` tinyint NOT NULL DEFAULT 0;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `messages` DROP `priority`;
-- +goose StatementEnd