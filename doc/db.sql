SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for files
-- ----------------------------
DROP TABLE IF EXISTS `files`;
CREATE TABLE `files`  (
                          `id` bigint NOT NULL AUTO_INCREMENT,
                          `filename` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
                          `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
                          `uploader` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
                          `uploader_id` bigint NULL DEFAULT NULL,
                          `link` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
                          `upload_time` bigint NULL DEFAULT NULL, -- 改为bigint
                          `download_counter` bigint NULL DEFAULT NULL,
                          PRIMARY KEY (`id`) USING BTREE,
                          UNIQUE INDEX `link`(`link` ASC) USING BTREE,
                          UNIQUE INDEX `link_2`(`link` ASC) USING BTREE,
                          INDEX `idx_filename`(`filename` ASC) USING BTREE,
                          INDEX `idx_uploader_id`(`uploader_id` ASC) USING BTREE,
                          INDEX `idx_link`(`link` ASC) USING BTREE,
                          INDEX `idx_files_uploader_id`(`uploader_id` ASC) USING BTREE,
                          INDEX `idx_files_link`(`link` ASC) USING BTREE,
                          INDEX `idx_files_filename`(`filename` ASC) USING BTREE,
                          INDEX `idx_files_uploader`(`uploader` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for outlines
-- ----------------------------
DROP TABLE IF EXISTS `outlines`;
CREATE TABLE `outlines`  (
                             `id` bigint NOT NULL AUTO_INCREMENT,
                             `project_id` bigint NULL DEFAULT NULL,
                             `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
                             `current_version` bigint NULL DEFAULT NULL,
                             `created_at` bigint NULL DEFAULT NULL, -- 改为bigint
                             `updated_at` bigint NULL DEFAULT NULL, -- 改为bigint
                             PRIMARY KEY (`id`) USING BTREE,
                             INDEX `idx_project_id`(`project_id` ASC) USING BTREE,
                             INDEX `idx_outlines_project_id`(`project_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for projects
-- ----------------------------
DROP TABLE IF EXISTS `projects`;
CREATE TABLE `projects`  (
                             `id` bigint NOT NULL AUTO_INCREMENT,
                             `user_id` bigint NULL DEFAULT NULL,
                             `username` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
                             `title` varchar(191) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
                             `description` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
                             `genre` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
                             `created_at` bigint NULL DEFAULT NULL, -- 改为bigint
                             `updated_at` bigint NULL DEFAULT NULL, -- 改为bigint
                             `last_edited_at` bigint NULL DEFAULT NULL, -- 改为bigint
                             PRIMARY KEY (`id`) USING BTREE,
                             INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
                             INDEX `idx_title`(`title` ASC) USING BTREE,
                             INDEX `idx_projects_title`(`title` ASC) USING BTREE,
                             INDEX `idx_projects_user_id`(`user_id` ASC) USING BTREE,
                             INDEX `idx_projects_username`(`username` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 3 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for referral_uses
-- ----------------------------
DROP TABLE IF EXISTS `referral_uses`;
CREATE TABLE `referral_uses`  (
                                  `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
                                  `referrer_id` bigint UNSIGNED NOT NULL,
                                  `user_id` bigint UNSIGNED NOT NULL,
                                  `referral_code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
                                  `tokens_rewarded` bigint NOT NULL,
                                  `used_at` bigint UNSIGNED NOT NULL, -- 改为bigint（存储Unix时间戳）
                                  `created_at` bigint UNSIGNED NULL DEFAULT NULL, -- 改为bigint
                                  PRIMARY KEY (`id`) USING BTREE,
                                  UNIQUE INDEX `user_id`(`user_id` ASC) USING BTREE,
                                  UNIQUE INDEX `user_id_2`(`user_id` ASC) USING BTREE,
                                  INDEX `idx_referrer_id`(`referrer_id` ASC) USING BTREE,
                                  INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
                                  INDEX `idx_referral_uses_referrer_id`(`referrer_id` ASC) USING BTREE,
                                  INDEX `idx_referral_uses_user_id`(`user_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for referrals
-- ----------------------------
DROP TABLE IF EXISTS `referrals`;
CREATE TABLE `referrals`  (
                              `id` bigint UNSIGNED NOT NULL AUTO_INCREMENT,
                              `user_id` bigint UNSIGNED NOT NULL,
                              `code` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
                              `total_used` bigint NULL DEFAULT 0,
                              `is_active` tinyint(1) NULL DEFAULT 1,
                              `created_at` bigint UNSIGNED NULL DEFAULT NULL, -- 改为bigint
                              `updated_at` bigint UNSIGNED NULL DEFAULT NULL, -- 改为bigint
                              PRIMARY KEY (`id`) USING BTREE,
                              UNIQUE INDEX `user_id`(`user_id` ASC) USING BTREE,
                              UNIQUE INDEX `code`(`code` ASC) USING BTREE,
                              UNIQUE INDEX `user_id_2`(`user_id` ASC) USING BTREE,
                              UNIQUE INDEX `code_2`(`code` ASC) USING BTREE,
                              INDEX `idx_code`(`code` ASC) USING BTREE,
                              INDEX `idx_user_id`(`user_id` ASC) USING BTREE,
                              INDEX `idx_referrals_user_id`(`user_id` ASC) USING BTREE,
                              INDEX `idx_referrals_code`(`code` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for versions
-- ----------------------------
DROP TABLE IF EXISTS `versions`;
CREATE TABLE `versions`  (
                             `id` bigint NOT NULL AUTO_INCREMENT,
                             `outline_id` bigint NULL DEFAULT NULL,
                             `version_number` bigint NULL DEFAULT NULL,
                             `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
                             `is_ai_generated` tinyint(1) NULL DEFAULT NULL,
                             `ai_style` longtext CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
                             `word_limit` bigint NULL DEFAULT NULL,
                             `tokens_used` bigint NULL DEFAULT NULL,
                             `created_at` bigint NULL DEFAULT NULL, -- 改为bigint
                             PRIMARY KEY (`id`) USING BTREE,
                             UNIQUE INDEX `unique_outline_version`(`outline_id` ASC, `version_number` ASC) USING BTREE,
                             INDEX `idx_outline_id`(`outline_id` ASC) USING BTREE,
                             INDEX `idx_versions_outline_id`(`outline_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- 其他表（options/packages/subscriptions等）的时间字段已默认是bigint类型，无需修改

SET FOREIGN_KEY_CHECKS = 1;