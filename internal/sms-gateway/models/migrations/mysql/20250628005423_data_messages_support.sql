-- +goose Up
-- +goose StatementBegin
ALTER TABLE `messages`
ADD `type` enum('Text', 'Data') NOT NULL DEFAULT 'Text';
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `messages`
SET `message` = json_object('text', `message`)
WHERE `is_hashed` = 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `messages` CHANGE `message` `content` text NOT NULL;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `messages` CHANGE `content` `message` text NOT NULL;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `messages`
SET `message` = COALESCE(
        json_value(`message`, '$.text'),
        json_value(`message`, '$.data')
    )
WHERE `is_hashed` = 0;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `messages` DROP `type`;
-- +goose StatementEnd