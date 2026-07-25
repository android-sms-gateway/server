-- +goose Up
-- +goose StatementBegin
CREATE TABLE `inbox` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `ext_id` VARCHAR(36) NOT NULL,
    `device_id` CHAR(21) NOT NULL,
    `type` ENUM('SMS', 'DATA_SMS', 'MMS', 'MMS_DOWNLOADED') NOT NULL,
    `sender` VARCHAR(512) NOT NULL,
    `recipient` VARCHAR(512) NULL,
    `sim_number` TINYINT(1) UNSIGNED NULL,
    `content` TEXT NOT NULL,
    `is_encrypted` TINYINT(1) UNSIGNED NOT NULL DEFAULT 1,
    `created_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3),
    `updated_at` DATETIME(3) NOT NULL DEFAULT CURRENT_TIMESTAMP(3) ON UPDATE CURRENT_TIMESTAMP(3),
    PRIMARY KEY (`id`),
    UNIQUE INDEX `unq_inbox_ext_device` (`ext_id`, `device_id`),
    INDEX `idx_inbox_device_created` (`device_id`, `created_at`),
    INDEX `idx_inbox_created` (`created_at`),
    CONSTRAINT `fk_inbox_device` FOREIGN KEY (`device_id`) REFERENCES `devices`(`id`) ON DELETE CASCADE
) ENGINE = InnoDB;
-- +goose StatementEnd
-- +goose StatementBegin
CREATE TABLE `inbox_attachments` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `message_id` BIGINT UNSIGNED NOT NULL,
    `part_id` BIGINT NOT NULL,
    `content_type` VARCHAR(128) NOT NULL,
    `name` VARCHAR(512) NOT NULL,
    `size` BIGINT NULL,
    `data` LONGBLOB NOT NULL,
    PRIMARY KEY (`id`),
    UNIQUE INDEX `unq_inbox_att_msg_part` (`message_id`, `part_id`),
    INDEX `idx_inbox_att_message` (`message_id`),
    CONSTRAINT `fk_inbox_att_message` FOREIGN KEY (`message_id`) REFERENCES `inbox`(`id`) ON DELETE CASCADE
) ENGINE = InnoDB;
-- +goose StatementEnd
---
-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS `inbox_attachments`;
-- +goose StatementEnd
-- +goose StatementBegin
DROP TABLE IF EXISTS `inbox`;
-- +goose StatementEnd