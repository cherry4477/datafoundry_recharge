CREATE TABLE IF NOT EXISTS DF_RECHARGE
(
    RECHARGE_ID       BIGINT NOT NULL AUTO_INCREMENT,
    AMOUNT            DOUBLE(13,2) NOT NULL,
    NAMESPACE         VARCHAR(128) NOT NULL,
    USER              VARCHAR(128) NOT NULL,
    CREATE_TIME       DATETIME,
    STATUS            VARCHAR(2) NOT NULL,
    STATUS_TIME       DATETIME,
    PRIMARY KEY (RECHARGE_ID)
)  DEFAULT CHARSET=UTF8;




CREATE TABLE IF NOT EXISTS `DF_balance` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `namespace` varchar(255) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,
  `created_at` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
  `updated_at` timestamp NULL DEFAULT NULL ON UPDATE CURRENT_TIMESTAMP,
  `balance` decimal(12,3) NOT NULL DEFAULT 0.000,
  PRIMARY KEY (`id`),
  UNIQUE KEY `namespace_unique` (`namespace`),
  KEY `df_balance_created_at_index` (`created_at`),
  KEY `df_balance_updated_at_index` (`updated_at`)
)  DEFAULT CHARSET=UTF8;

