-- Migration: init
-- Created at: 2017-10-07 13:37:25
-- ====  UP  ====

BEGIN;

CREATE TABLE IF NOT EXISTS `crons` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,

    `user_id` INT NOT NULL,
    `query` TEXT NOT NULL,
    `sources` TEXT NOT NULL,

    `created_at` DATETIME(6) NOT NULL,
    `updated_at` DATETIME(6) NOT NULL,

    PRIMARY KEY (`id`),
    INDEX `cron_user_id_idx` (`user_id`)
)
ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

COMMIT;

-- ==== DOWN ====

BEGIN;

DROP TABLE IF EXISTS `crons`;

COMMIT;
