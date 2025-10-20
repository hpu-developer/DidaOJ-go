/*
 Navicat Premium Dump SQL

 Source Server         : 1panel.didapipa.com_didaoj
 Source Server Type    : MySQL
 Source Server Version : 80406 (8.4.6)
 Source Host           : 1panel.didapipa.com:3306
 Source Schema         : didaoj

 Target Server Type    : MySQL
 Target Server Version : 80406 (8.4.6)
 File Encoding         : 65001

 Date: 20/10/2025 18:56:23
*/

SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;

-- ----------------------------
-- Table structure for collection
-- ----------------------------
DROP TABLE IF EXISTS `collection`;
CREATE TABLE `collection`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `start_time` datetime NULL DEFAULT NULL,
  `end_time` datetime NULL DEFAULT NULL,
  `inserter` int NULL DEFAULT NULL,
  `insert_time` datetime NULL DEFAULT NULL,
  `modifier` int NULL DEFAULT NULL,
  `modify_time` datetime NULL DEFAULT NULL,
  `private` tinyint(1) NULL DEFAULT NULL,
  `password` varchar(30) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 16 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for collection_member
-- ----------------------------
DROP TABLE IF EXISTS `collection_member`;
CREATE TABLE `collection_member`  (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`, `user_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '题集参与人员' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for collection_problem
-- ----------------------------
DROP TABLE IF EXISTS `collection_problem`;
CREATE TABLE `collection_problem`  (
  `id` int NOT NULL,
  `problem_id` int NOT NULL,
  `index` int NULL DEFAULT NULL,
  PRIMARY KEY (`id`, `problem_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '比赛限制语言' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for contest
-- ----------------------------
DROP TABLE IF EXISTS `contest`;
CREATE TABLE `contest`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(75) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  `notification` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `start_time` datetime NULL DEFAULT NULL,
  `end_time` datetime NULL DEFAULT NULL,
  `inserter` int NULL DEFAULT NULL,
  `insert_time` datetime NULL DEFAULT NULL,
  `modifier` int NULL DEFAULT NULL,
  `modify_time` datetime NULL DEFAULT NULL,
  `private` tinyint(1) NULL DEFAULT NULL,
  `password` varchar(35) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `submit_anytime` tinyint(1) NULL DEFAULT NULL,
  `type` tinyint NULL DEFAULT NULL,
  `score_type` tinyint NULL DEFAULT NULL,
  `lock_rank_duration` bigint NULL DEFAULT NULL,
  `always_lock` tinyint(1) NULL DEFAULT NULL,
  `discuss_type` tinyint NULL DEFAULT NULL COMMENT '讨论类型，0正常讨论，1仅查看自己的讨论',
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 105 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for contest_language
-- ----------------------------
DROP TABLE IF EXISTS `contest_language`;
CREATE TABLE `contest_language`  (
  `id` int NOT NULL,
  `language` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '比赛限制语言' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for contest_member
-- ----------------------------
DROP TABLE IF EXISTS `contest_member`;
CREATE TABLE `contest_member`  (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`, `user_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '比赛访问权限' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for contest_member_auth
-- ----------------------------
DROP TABLE IF EXISTS `contest_member_auth`;
CREATE TABLE `contest_member_auth`  (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`, `user_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '比赛管理权限' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for contest_member_author
-- ----------------------------
DROP TABLE IF EXISTS `contest_member_author`;
CREATE TABLE `contest_member_author`  (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`, `user_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '比赛作者列表' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for contest_member_ignore
-- ----------------------------
DROP TABLE IF EXISTS `contest_member_ignore`;
CREATE TABLE `contest_member_ignore`  (
  `user_id` int NOT NULL,
  `id` int NOT NULL,
  PRIMARY KEY (`id`, `user_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '比赛忽略排行成员' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for contest_member_volunteer
-- ----------------------------
DROP TABLE IF EXISTS `contest_member_volunteer`;
CREATE TABLE `contest_member_volunteer`  (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`, `user_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '比赛志愿者权限' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for contest_problem
-- ----------------------------
DROP TABLE IF EXISTS `contest_problem`;
CREATE TABLE `contest_problem`  (
  `id` int NOT NULL,
  `problem_id` int NULL DEFAULT NULL,
  `index` int NULL DEFAULT NULL,
  `view_id` int NULL DEFAULT NULL,
  `score` int NULL DEFAULT NULL,
  UNIQUE INDEX `group_id`(`id` ASC, `problem_id` ASC) USING BTREE,
  UNIQUE INDEX `group_index`(`id` ASC, `index` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '比赛限制语言' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for discuss
-- ----------------------------
DROP TABLE IF EXISTS `discuss`;
CREATE TABLE `discuss`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `content` mediumtext CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `view_count` int NOT NULL,
  `banned` tinyint(1) NULL DEFAULT NULL,
  `problem_id` int NULL DEFAULT NULL,
  `contest_id` int NULL DEFAULT NULL,
  `judge_id` int NULL DEFAULT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL,
  `modifier` int NOT NULL,
  `modify_time` datetime NOT NULL,
  `updater` int NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `discuss_pk_2`(`id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 169 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for discuss_comment
-- ----------------------------
DROP TABLE IF EXISTS `discuss_comment`;
CREATE TABLE `discuss_comment`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `discuss_id` int NOT NULL,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `banned` tinyint(1) NOT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL,
  `modifier` int NOT NULL,
  `modify_time` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `discuss_pk`(`id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 81 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for discuss_tag
-- ----------------------------
DROP TABLE IF EXISTS `discuss_tag`;
CREATE TABLE `discuss_tag`  (
  `id` int NOT NULL,
  `tag_id` int NOT NULL,
  `index` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`, `tag_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for judge_job
-- ----------------------------
DROP TABLE IF EXISTS `judge_job`;
CREATE TABLE `judge_job`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `problem_id` int NOT NULL,
  `contest_id` int NULL DEFAULT NULL,
  `language` tinyint(1) NOT NULL,
  `code` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `code_length` int NOT NULL,
  `status` tinyint(1) NOT NULL,
  `judger` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `judge_time` datetime NULL DEFAULT NULL,
  `task_current` int NULL DEFAULT NULL,
  `task_total` int NULL DEFAULT NULL,
  `score` int NULL DEFAULT NULL,
  `time` bigint NULL DEFAULT NULL,
  `memory` bigint NULL DEFAULT NULL,
  `private` tinyint(1) NOT NULL,
  `remote_judge_id` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `remote_account_id` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  INDEX `idx_contest_status_inserter_problem_id_id`(`contest_id` ASC, `status` ASC, `inserter` ASC, `problem_id` ASC, `id` ASC) USING BTREE,
  INDEX `idx_contest_inserter_problem_id_id`(`contest_id` ASC, `inserter` ASC, `problem_id` ASC, `id` ASC) USING BTREE,
  INDEX `idx_status_id`(`status` ASC, `id` ASC) USING BTREE,
  INDEX `idx_status_problem_id`(`status` ASC, `problem_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 150846 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for judge_job_compile
-- ----------------------------
DROP TABLE IF EXISTS `judge_job_compile`;
CREATE TABLE `judge_job_compile`  (
  `id` int NULL DEFAULT NULL,
  `message` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL,
  UNIQUE INDEX `judge_job_compile_pk`(`id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for judge_task
-- ----------------------------
DROP TABLE IF EXISTS `judge_task`;
CREATE TABLE `judge_task`  (
  `id` int NOT NULL,
  `task_id` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `status` tinyint(1) NULL DEFAULT NULL,
  `time` bigint NULL DEFAULT NULL,
  `memory` bigint NULL DEFAULT NULL,
  `score` int NULL DEFAULT NULL,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
  `hint` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL,
  PRIMARY KEY (`id`, `task_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_general_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for judger
-- ----------------------------
DROP TABLE IF EXISTS `judger`;
CREATE TABLE `judger`  (
  `key` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `max_job` int NULL DEFAULT NULL,
  `cpu_usage` float NULL DEFAULT NULL,
  `mem_usage` bigint UNSIGNED NULL DEFAULT NULL,
  `mem_total` bigint UNSIGNED NULL DEFAULT NULL,
  `avg_message` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `insert_time` datetime NULL DEFAULT NULL,
  `modify_time` datetime NULL DEFAULT NULL,
  `hidden` tinyint(1) NULL DEFAULT NULL,
  `enable` tinyint(1) NULL DEFAULT NULL COMMENT '是否启用',
  PRIMARY KEY (`key`) USING BTREE,
  INDEX `judger_enable_index`(`enable` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for problem
-- ----------------------------
DROP TABLE IF EXISTS `problem`;
CREATE TABLE `problem`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '问题ID',
  `key` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NULL DEFAULT NULL,
  `title` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '标题',
  `description` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '描述',
  `source` varchar(200) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `time_limit` int NOT NULL COMMENT '时间限制',
  `memory_limit` int NOT NULL COMMENT '内存限制',
  `judge_type` tinyint(1) NOT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL COMMENT '创建时间',
  `modifier` int NOT NULL,
  `modify_time` datetime NOT NULL COMMENT '创建时间',
  `accept` int NOT NULL COMMENT '通过人数',
  `attempt` int NOT NULL COMMENT '尝试人数',
  `private` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `problem_pk`(`key` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 2140 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for problem_daily
-- ----------------------------
DROP TABLE IF EXISTS `problem_daily`;
CREATE TABLE `problem_daily`  (
  `key` char(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `problem_id` int NOT NULL,
  `solution` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `code` text CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `inserter` int NULL DEFAULT NULL,
  `modifier` int NULL DEFAULT NULL,
  `insert_time` datetime NULL DEFAULT NULL,
  `modify_time` datetime NULL DEFAULT NULL,
  PRIMARY KEY (`key`) USING BTREE,
  UNIQUE INDEX `problem_daily_pk_2`(`problem_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for problem_local
-- ----------------------------
DROP TABLE IF EXISTS `problem_local`;
CREATE TABLE `problem_local`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '问题ID',
  `problem_id` int NOT NULL,
  `judge_md5` char(32) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `problem_local_pk`(`problem_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 1808 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for problem_member
-- ----------------------------
DROP TABLE IF EXISTS `problem_member`;
CREATE TABLE `problem_member`  (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`, `user_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '问题访问权限' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for problem_member_auth
-- ----------------------------
DROP TABLE IF EXISTS `problem_member_auth`;
CREATE TABLE `problem_member_auth`  (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`, `user_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci COMMENT = '问题管理权限' ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for problem_remote
-- ----------------------------
DROP TABLE IF EXISTS `problem_remote`;
CREATE TABLE `problem_remote`  (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '问题ID',
  `problem_id` int NOT NULL,
  `origin_oj` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '原始OJ标识',
  `origin_id` varchar(8) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '原始ID标识',
  `origin_url` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '原始链接',
  `origin_author` varchar(255) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `problem_remote_pk`(`problem_id` ASC) USING BTREE,
  UNIQUE INDEX `problem_origin`(`origin_oj` ASC, `origin_id` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 333 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for problem_tag
-- ----------------------------
DROP TABLE IF EXISTS `problem_tag`;
CREATE TABLE `problem_tag`  (
  `id` int NOT NULL,
  `tag_id` int NOT NULL,
  `index` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`, `tag_id`) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for tag
-- ----------------------------
DROP TABLE IF EXISTS `tag`;
CREATE TABLE `tag`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci NULL DEFAULT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `tag_pk`(`name` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 190 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS `user`;
CREATE TABLE `user`  (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '用户名',
  `nickname` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL COMMENT '昵称',
  `real_name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '真实名称',
  `password` char(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `email` varchar(90) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `gender` tinyint(1) NULL DEFAULT NULL COMMENT '性别',
  `number` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '身份标识',
  `slogan` varchar(50) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '签名',
  `organization` varchar(80) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT '组织',
  `qq` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'QQ',
  `vjudge_id` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'VjudgeId',
  `github` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL COMMENT 'Github',
  `codeforces` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  `check_in_count` int NOT NULL COMMENT '签到次数',
  `insert_time` datetime NOT NULL,
  `modify_time` datetime NOT NULL,
  `accept` int NOT NULL,
  `attempt` int NOT NULL,
  PRIMARY KEY (`id`) USING BTREE,
  UNIQUE INDEX `uk_username`(`username` ASC) USING BTREE
) ENGINE = InnoDB AUTO_INCREMENT = 6623 CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

-- ----------------------------
-- Table structure for user_role
-- ----------------------------
DROP TABLE IF EXISTS `user_role`;
CREATE TABLE `user_role`  (
  `id` int NULL DEFAULT NULL,
  `role_id` varchar(10) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NULL DEFAULT NULL,
  UNIQUE INDEX `group_id`(`id` ASC, `role_id` ASC) USING BTREE
) ENGINE = InnoDB CHARACTER SET = utf8mb4 COLLATE = utf8mb4_0900_ai_ci ROW_FORMAT = Dynamic;

SET FOREIGN_KEY_CHECKS = 1;
