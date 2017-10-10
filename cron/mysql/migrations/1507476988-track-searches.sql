-- Migration: track-searches
-- Created at: 2017-10-08 16:36:28
-- ====  UP  ====

BEGIN;

CREATE TABLE IF NOT EXISTS `search_results` (
    `id` INT UNSIGNED NOT NULL AUTO_INCREMENT,

    `cron_id` INT UNSIGNED NOT NULL,
    `source` VARCHAR(256) NOT NULL,

    `result` TEXT NOT NULL,

    `created_at` DATETIME NOT NULL,

    PRIMARY KEY (`id`),
    CONSTRAINT `fk_search_results_crons` FOREIGN KEY (`cron_id`) REFERENCES `crons` (`id`) ON DELETE CASCADE,
    INDEX `search_results_cron_id_source_created_at_INDEX` (`cron_id`, `source`, `created_at`)
)
ENGINE=InnoDB DEFAULT CHARSET=utf8 COLLATE=utf8_unicode_ci;

COMMIT;

-- ==== DOWN ====

BEGIN;

DROP TABLE IF EXISTS `search_results`;

COMMIT;

