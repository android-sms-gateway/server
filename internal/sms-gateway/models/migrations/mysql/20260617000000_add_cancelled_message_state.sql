-- +goose Up
-- +goose StatementBegin
ALTER TABLE `messages`
MODIFY COLUMN `state` enum(
        'Pending',
        'Cancelling',
        'Cancelled',
        'Processed',
        'Sent',
        'Delivered',
        'Failed'
    ) NOT NULL DEFAULT 'Pending';
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `message_recipients`
MODIFY COLUMN `state` enum(
        'Pending',
        'Cancelling',
        'Cancelled',
        'Processed',
        'Sent',
        'Delivered',
        'Failed'
    ) NOT NULL DEFAULT 'Pending';
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `message_states`
MODIFY COLUMN `state` enum(
        'Pending',
        'Cancelling',
        'Cancelled',
        'Processed',
        'Sent',
        'Delivered',
        'Failed'
    ) NOT NULL;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
UPDATE `messages`
SET `state` = CASE
        WHEN `state` = 'Cancelling' THEN 'Pending'
        WHEN `state` = 'Cancelled' THEN 'Failed'
        ELSE `state`
    END
WHERE `state` IN ('Cancelling', 'Cancelled');
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `messages`
MODIFY COLUMN `state` enum(
        'Pending',
        'Processed',
        'Sent',
        'Delivered',
        'Failed'
    ) NOT NULL DEFAULT 'Pending';
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `message_recipients`
SET `state` = CASE
        WHEN `state` = 'Cancelling' THEN 'Pending'
        WHEN `state` = 'Cancelled' THEN 'Failed'
        ELSE `state`
    END
WHERE `state` IN ('Cancelling', 'Cancelled');
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `message_recipients`
MODIFY COLUMN `state` enum(
        'Pending',
        'Processed',
        'Sent',
        'Delivered',
        'Failed'
    ) NOT NULL DEFAULT 'Pending';
-- +goose StatementEnd
-- +goose StatementBegin
UPDATE `message_states`
SET `state` = CASE
        WHEN `state` = 'Cancelling' THEN 'Pending'
        WHEN `state` = 'Cancelled' THEN 'Failed'
        ELSE `state`
    END
WHERE `state` IN ('Cancelling', 'Cancelled');
-- +goose StatementEnd
-- +goose StatementBegin
ALTER TABLE `message_states`
MODIFY COLUMN `state` enum(
        'Pending',
        'Processed',
        'Sent',
        'Delivered',
        'Failed'
    ) NOT NULL;
-- +goose StatementEnd