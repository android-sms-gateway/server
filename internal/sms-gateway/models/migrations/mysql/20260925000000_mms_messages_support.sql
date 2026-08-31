-- +goose Up
-- +goose StatementBegin
ALTER TABLE `messages`
MODIFY COLUMN `content` mediumtext NOT NULL,
    MODIFY COLUMN `type` enum('Text', 'Data', 'Mms') NOT NULL DEFAULT 'Text';
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DELETE FROM `messages`
WHERE `type` = 'Mms';
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `messages`
MODIFY COLUMN `content` text NOT NULL,
    MODIFY COLUMN `type` enum('Text', 'Data') NOT NULL DEFAULT 'Text';
-- +goose StatementEnd