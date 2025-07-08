-- +goose Up
-- +goose StatementBegin
ALTER TABLE `messages` DROP `message`;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `messages`
ADD `message` text NULL;
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `messages`
SET `message` = CASE
        WHEN `is_hashed` = 0 THEN COALESCE(
            JSON_VALUE(`content`, '$.text'),
            JSON_VALUE(`content`, '$.data')
        )
        ELSE `content`
    END;
-- +goose StatementEnd