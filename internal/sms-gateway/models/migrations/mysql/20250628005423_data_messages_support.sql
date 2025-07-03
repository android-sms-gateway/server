-- +goose Up
-- +goose StatementBegin
ALTER TABLE `messages`
ADD `type` enum('Text', 'Data') NOT NULL DEFAULT 'Text',
    ADD `content` text NOT NULL,
    MODIFY `message` text NULL;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `messages`
SET `content` = json_object('text', `message`)
WHERE `is_hashed` = 0;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `messages`
SET `content` = `message`
WHERE `is_hashed` = 1;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
UPDATE `messages`
SET `message` = COALESCE(
        json_value(`content`, '$.text'),
        json_value(`content`, '$.data')
    )
WHERE `is_hashed` = 0;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `messages`
SET `message` = `content`
WHERE `is_hashed` = 1;
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `messages` DROP `type`,
    DROP `content`,
    MODIFY `message` text NOT NULL;
-- +goose StatementEnd