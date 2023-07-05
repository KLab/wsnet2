
DROP TABLE IF EXISTS `game_server`;
CREATE TABLE `game_server` (
  `id`          INTEGER UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `hostname`    VARCHAR(191) NOT NULL,
  `public_name` VARCHAR(191) NOT NULL,
  `grpc_port`   INTEGER NOT NULL,
  `ws_port`     INTEGER NOT NULL,
  `status`      TINYINT NOT NULL,
  `heartbeat`   BIGINT,
  UNIQUE KEY `idx_hostname` (`hostname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `hub_server`;
CREATE TABLE `hub_server` (
  `id`          INTEGER UNSIGNED NOT NULL PRIMARY KEY AUTO_INCREMENT,
  `hostname`    VARCHAR(191) NOT NULL,
  `public_name` VARCHAR(191) NOT NULL,
  `grpc_port`   INTEGER NOT NULL,
  `ws_port`     INTEGER NOT NULL,
  `status`      TINYINT NOT NULL,
  `heartbeat`   BIGINT,
  UNIQUE KEY `idx_hostname` (`hostname`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `app`;
CREATE TABLE app (
  `id`   VARCHAR(32) COLLATE ascii_bin PRIMARY KEY,
  `name` VARCHAR(191) COLLATE utf8mb4_bin,
  `key`  VARCHAR(191) COLLATE ascii_bin
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `room`;
CREATE TABLE room (
  `id`     VARCHAR(32) PRIMARY KEY,
  `app_id` VARCHAR(32) NOT NULL,
  `host_id` INTEGER UNSIGNED NOT NULL,
  `visible` TINYINT NOT NULL,
  `joinable` TINYINT NOT NULL,
  `watchable` TINYINT NOT NULL,
  `number` INTEGER,
  `search_group` INTEGER UNSIGNED NOT NULL,
  `max_players` INTEGER UNSIGNED NOT NULL,
  `players` INTEGER UNSIGNED NOT NULL,
  `watchers` INTEGER UNSIGNED NOT NULL,
  `props` BLOB,
  `created` DATETIME,
  UNIQUE KEY `idx_number` (`number`),
  KEY `idx_search_group` (`app_id`, `search_group`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `room_history`;
CREATE TABLE `room_history` (
  `id` BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `app_id` VARCHAR(32) NOT NULL,
  `host_id` INTEGER UNSIGNED NOT NULL,
  `room_id` VARCHAR(32) NOT NULL,
  `number` INTEGER,
  `search_group` INTEGER UNSIGNED NOT NULL,
  `max_players` INTEGER UNSIGNED NOT NULL,
  `public_props` BLOB,
  `private_props` BLOB,
  `created` DATETIME,
  `closed` DATETIME,
  KEY `room_id` (`room_id`),
  KEY `created` (`created`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `player_log`;
CREATE TABLE player_log (
  `id`        BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `room_id`   VARCHAR(32) NOT NULL,
  `player_id` VARCHAR(32) NOT NULL,
  `message`   VARCHAR(32) NOT NULL,
  `datetime`  DATETIME,
  KEY `room_id` (`room_id`),
  KEY `player_id` (`player_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;

DROP TABLE IF EXISTS `hub`;
CREATE TABLE hub (
  `id`      BIGINT UNSIGNED PRIMARY KEY AUTO_INCREMENT,
  `host_id` INTEGER UNSIGNED NOT NULL,
  `room_id` VARCHAR(32) NOT NULL,
  `watchers` INTEGER UNSIGNED NOT NULL,
  `created` DATETIME NOT NULL,
  UNIQUE KEY `idx_room` (`room_id`, `host_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4;
