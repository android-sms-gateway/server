-- +goose Up
-- +goose StatementBegin
ALTER TABLE `tokens`
ADD COLUMN `token_use` ENUM('access', 'refresh') NOT NULL DEFAULT 'access',
    ADD COLUMN `parent_jti` char(21) DEFAULT NULL,
    ADD INDEX `idx_tokens_parent_jti` (`parent_jti`);
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
ALTER TABLE `tokens` DROP INDEX `idx_tokens_parent_jti`,
    DROP COLUMN `token_use`,
    DROP COLUMN `parent_jti`;
-- +goose StatementEnd