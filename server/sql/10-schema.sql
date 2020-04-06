
DROP TABLE IF EXISTS `app`;
CREATE TABLE app (
  `id`   VARCHAR(32) COLLATE ascii_bin PRIMARY KEY,
  `name` VARCHAR(255) COLLATE utf8mb4_bin,
  `key`  VARCHAR(255) COLLATE ascii_bin
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `room`;
CREATE TABLE room (
  `id`     VARCHAR(32) PRIMARY KEY,
  `app_id` VARCHAR(32) NOT NULL,
  `host_id` INTEGER UNSIGNED NOT NULL,
  `visible` TINYINT NOT NULL,
  `joinable` TINYINT NOT NULL,
  `watchable` TINYINT NOT NULL,
  `number` INTEGER NOT NULL,
  `search_group` INTEGER UNSIGNED NOT NULL,
  `max_players` INTEGER UNSIGNED NOT NULL,
  `props` BLOB,
  `created` TIMESTAMP,
  KEY `idx_search_group` (`app_id`, `search_group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
