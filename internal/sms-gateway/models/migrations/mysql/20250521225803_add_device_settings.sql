-- +goose Up
-- +goose StatementBegin
CREATE TABLE `device_settings` (
    `user_id` varchar(32) NOT NULL,
    `settings` json NOT NULL,
    PRIMARY KEY (`user_id`),
    CONSTRAINT `fk_device_settings_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP TABLE `device_settings`;
-- +goose StatementEnd