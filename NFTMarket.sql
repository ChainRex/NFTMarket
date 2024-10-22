-- Create syntax for TABLE 'nft_attributes'
CREATE TABLE `nft_attributes` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `nft_id` bigint unsigned NOT NULL,
  `trait_type` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `value` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `nft_id_trait_type` (`trait_type`,`nft_id`),
  KEY `idx_nft_attributes_nft_id` (`nft_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create syntax for TABLE 'nft_collections'
CREATE TABLE `nft_collections` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `contract_address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `symbol` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `token_icon_uri` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `contract_address` (`contract_address`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create syntax for TABLE 'nft_transfer_events'
CREATE TABLE `nft_transfer_events` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `contract_address` varchar(42) COLLATE utf8mb4_unicode_ci NOT NULL,
  `token_id` bigint unsigned NOT NULL,
  `event_type` enum('mint','transfer') COLLATE utf8mb4_unicode_ci NOT NULL,
  `from_address` varchar(42) COLLATE utf8mb4_unicode_ci NOT NULL,
  `to_address` varchar(42) COLLATE utf8mb4_unicode_ci NOT NULL,
  `transaction_hash` varchar(66) COLLATE utf8mb4_unicode_ci NOT NULL,
  `block_number` bigint unsigned NOT NULL,
  `block_timestamp` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_contract_token` (`contract_address`,`token_id`),
  KEY `idx_from` (`from_address`),
  KEY `idx_to` (`to_address`),
  KEY `idx_block_number` (`block_number`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create syntax for TABLE 'nfts'
CREATE TABLE `nfts` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `collection_id` bigint unsigned NOT NULL,
  `token_id` bigint unsigned NOT NULL,
  `contract_address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `owner` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `token_uri` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `name` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `image` text CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `idx_collection_token` (`contract_address`,`token_id`),
  KEY `idx_nfts_owner` (`owner`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;

-- Create syntax for TABLE 'orders'
CREATE TABLE `orders` (
  `id` bigint unsigned NOT NULL AUTO_INCREMENT,
  `nft_contract_address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `token_id` bigint unsigned NOT NULL,
  `token_address` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `price` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `seller` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NOT NULL,
  `status` tinyint unsigned NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_orders_nft_contract_address` (`nft_contract_address`),
  KEY `idx_orders_token_id` (`token_id`),
  KEY `idx_orders_seller` (`seller`)
) ENGINE=InnoDB  DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;