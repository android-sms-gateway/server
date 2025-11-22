-- +goose Up
-- +goose StatementBegin
CREATE TABLE `tokens` (
    `id` char(21) NOT NULL PRIMARY KEY,
    `user_id` char(21) NOT NULL,
    `expires_at` datetime(3) NOT NULL,
    `created_at` datetime(3) NOT NULL DEFAULT current_timestamp(3),
    `updated_at` datetime(3) NOT NULL DEFAULT current_timestamp(3) ON UPDATE current_timestamp(3),
    `revoked_at` datetime(3) NULL,
    INDEX `idx_tokens_user_id` (`user_id`),
    INDEX `idx_tokens_expires_at` (`expires_at`),
    CONSTRAINT `fk_tokens_user` FOREIGN KEY (`user_id`) REFERENCES `users`(`id`) ON DELETE CASCADE
);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP TABLE `tokens`;
-- +goose StatementEnd