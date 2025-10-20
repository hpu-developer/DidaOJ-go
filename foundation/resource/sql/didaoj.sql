-- MySQL dump 10.13  Distrib 9.4.0, for Win64 (x86_64)
--
-- Host: 59.110.20.167    Database: didaoj
-- ------------------------------------------------------
-- Server version	8.4.6

/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!50503 SET NAMES utf8mb4 */;
/*!40103 SET @OLD_TIME_ZONE=@@TIME_ZONE */;
/*!40103 SET TIME_ZONE='+00:00' */;
/*!40014 SET @OLD_UNIQUE_CHECKS=@@UNIQUE_CHECKS, UNIQUE_CHECKS=0 */;
/*!40014 SET @OLD_FOREIGN_KEY_CHECKS=@@FOREIGN_KEY_CHECKS, FOREIGN_KEY_CHECKS=0 */;
/*!40101 SET @OLD_SQL_MODE=@@SQL_MODE, SQL_MODE='NO_AUTO_VALUE_ON_ZERO' */;
/*!40111 SET @OLD_SQL_NOTES=@@SQL_NOTES, SQL_NOTES=0 */;

--
-- Table structure for table `collection`
--

DROP TABLE IF EXISTS `collection`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `collection` (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(30) DEFAULT NULL,
  `description` text,
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `inserter` int DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modifier` int DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  `private` tinyint(1) DEFAULT NULL,
  `password` varchar(30) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=16 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `collection_member`
--

DROP TABLE IF EXISTS `collection_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `collection_member` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='题集参与人员';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `collection_problem`
--

DROP TABLE IF EXISTS `collection_problem`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `collection_problem` (
  `id` int NOT NULL,
  `problem_id` int NOT NULL,
  `index` int DEFAULT NULL,
  PRIMARY KEY (`id`,`problem_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛限制语言';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `contest`
--

DROP TABLE IF EXISTS `contest`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest` (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(75) NOT NULL,
  `description` text,
  `notification` varchar(100) DEFAULT NULL,
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `inserter` int DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modifier` int DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  `private` tinyint(1) DEFAULT NULL,
  `password` varchar(35) DEFAULT NULL,
  `submit_anytime` tinyint(1) DEFAULT NULL,
  `type` tinyint DEFAULT NULL,
  `score_type` tinyint DEFAULT NULL,
  `lock_rank_duration` bigint DEFAULT NULL,
  `always_lock` tinyint(1) DEFAULT NULL,
  `discuss_type` tinyint DEFAULT NULL COMMENT '讨论类型，0正常讨论，1仅查看自己的讨论',
  PRIMARY KEY (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=105 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `contest_language`
--

DROP TABLE IF EXISTS `contest_language`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest_language` (
  `id` int NOT NULL,
  `language` varchar(10) DEFAULT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛限制语言';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `contest_member`
--

DROP TABLE IF EXISTS `contest_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest_member` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛访问权限';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `contest_member_auth`
--

DROP TABLE IF EXISTS `contest_member_auth`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest_member_auth` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛管理权限';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `contest_member_author`
--

DROP TABLE IF EXISTS `contest_member_author`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest_member_author` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛作者列表';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `contest_member_ignore`
--

DROP TABLE IF EXISTS `contest_member_ignore`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest_member_ignore` (
  `user_id` int NOT NULL,
  `id` int NOT NULL,
  PRIMARY KEY (`id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛忽略排行成员';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `contest_member_volunteer`
--

DROP TABLE IF EXISTS `contest_member_volunteer`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest_member_volunteer` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛志愿者权限';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `contest_problem`
--

DROP TABLE IF EXISTS `contest_problem`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `contest_problem` (
  `id` int NOT NULL,
  `problem_id` int DEFAULT NULL,
  `index` int DEFAULT NULL,
  `view_id` int DEFAULT NULL,
  `score` int DEFAULT NULL,
  UNIQUE KEY `group_id` (`id`,`problem_id`),
  UNIQUE KEY `group_index` (`id`,`index`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛限制语言';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `discuss`
--

DROP TABLE IF EXISTS `discuss`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `discuss` (
  `id` int NOT NULL AUTO_INCREMENT,
  `title` varchar(100) NOT NULL,
  `content` mediumtext NOT NULL,
  `view_count` int NOT NULL,
  `banned` tinyint(1) DEFAULT NULL,
  `problem_id` int DEFAULT NULL,
  `contest_id` int DEFAULT NULL,
  `judge_id` int DEFAULT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL,
  `modifier` int NOT NULL,
  `modify_time` datetime NOT NULL,
  `updater` int NOT NULL,
  `update_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `discuss_pk_2` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=169 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `discuss_comment`
--

DROP TABLE IF EXISTS `discuss_comment`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `discuss_comment` (
  `id` int NOT NULL AUTO_INCREMENT,
  `discuss_id` int NOT NULL,
  `content` text NOT NULL,
  `banned` tinyint(1) NOT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL,
  `modifier` int NOT NULL,
  `modify_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `discuss_pk` (`id`)
) ENGINE=InnoDB AUTO_INCREMENT=81 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `discuss_tag`
--

DROP TABLE IF EXISTS `discuss_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `discuss_tag` (
  `id` int NOT NULL,
  `tag_id` int NOT NULL,
  `index` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`,`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `judge_job`
--

DROP TABLE IF EXISTS `judge_job`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `judge_job` (
  `id` int NOT NULL AUTO_INCREMENT,
  `problem_id` int NOT NULL,
  `contest_id` int DEFAULT NULL,
  `language` tinyint(1) NOT NULL,
  `code` text NOT NULL,
  `code_length` int NOT NULL,
  `status` tinyint(1) NOT NULL,
  `judger` varchar(10) DEFAULT NULL,
  `judge_time` datetime DEFAULT NULL,
  `task_current` int DEFAULT NULL,
  `task_total` int DEFAULT NULL,
  `score` int DEFAULT NULL,
  `time` bigint DEFAULT NULL,
  `memory` bigint DEFAULT NULL,
  `private` tinyint(1) NOT NULL,
  `remote_judge_id` varchar(20) DEFAULT NULL,
  `remote_account_id` varchar(20) DEFAULT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL,
  PRIMARY KEY (`id`),
  KEY `idx_contest_status_inserter_problem_id_id` (`contest_id`,`status`,`inserter`,`problem_id`,`id`),
  KEY `idx_contest_inserter_problem_id_id` (`contest_id`,`inserter`,`problem_id`,`id`),
  KEY `idx_status_id` (`status`,`id`),
  KEY `idx_status_problem_id` (`status`,`problem_id`)
) ENGINE=InnoDB AUTO_INCREMENT=150845 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `judge_job_compile`
--

DROP TABLE IF EXISTS `judge_job_compile`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `judge_job_compile` (
  `id` int DEFAULT NULL,
  `message` text,
  UNIQUE KEY `judge_job_compile_pk` (`id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `judge_task`
--

DROP TABLE IF EXISTS `judge_task`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `judge_task` (
  `id` int NOT NULL,
  `task_id` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci NOT NULL,
  `status` tinyint(1) DEFAULT NULL,
  `time` bigint DEFAULT NULL,
  `memory` bigint DEFAULT NULL,
  `score` int DEFAULT NULL,
  `content` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  `hint` text CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci,
  PRIMARY KEY (`id`,`task_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `judger`
--

DROP TABLE IF EXISTS `judger`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `judger` (
  `key` varchar(10) NOT NULL,
  `name` varchar(20) NOT NULL,
  `max_job` int DEFAULT NULL,
  `cpu_usage` float DEFAULT NULL,
  `mem_usage` bigint unsigned DEFAULT NULL,
  `mem_total` bigint unsigned DEFAULT NULL,
  `avg_message` varchar(100) DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  `hidden` tinyint(1) DEFAULT NULL,
  `enable` tinyint(1) DEFAULT NULL COMMENT '是否启用',
  PRIMARY KEY (`key`),
  KEY `judger_enable_index` (`enable`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `problem`
--

DROP TABLE IF EXISTS `problem`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '问题ID',
  `key` varchar(15) CHARACTER SET utf8mb4 COLLATE utf8mb4_general_ci DEFAULT NULL,
  `title` varchar(50) NOT NULL COMMENT '标题',
  `description` text NOT NULL COMMENT '描述',
  `source` varchar(200) DEFAULT NULL,
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
  PRIMARY KEY (`id`),
  UNIQUE KEY `problem_pk` (`key`)
) ENGINE=InnoDB AUTO_INCREMENT=2140 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `problem_daily`
--

DROP TABLE IF EXISTS `problem_daily`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_daily` (
  `key` char(10) NOT NULL,
  `problem_id` int NOT NULL,
  `solution` text NOT NULL,
  `code` text NOT NULL,
  `inserter` int DEFAULT NULL,
  `modifier` int DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  PRIMARY KEY (`key`),
  UNIQUE KEY `problem_daily_pk_2` (`problem_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `problem_local`
--

DROP TABLE IF EXISTS `problem_local`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_local` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '问题ID',
  `problem_id` int NOT NULL,
  `judge_md5` char(32) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `problem_local_pk` (`problem_id`)
) ENGINE=InnoDB AUTO_INCREMENT=1808 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `problem_member`
--

DROP TABLE IF EXISTS `problem_member`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_member` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='问题访问权限';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `problem_member_auth`
--

DROP TABLE IF EXISTS `problem_member_auth`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_member_auth` (
  `id` int NOT NULL,
  `user_id` int NOT NULL,
  PRIMARY KEY (`id`,`user_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='问题管理权限';
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `problem_remote`
--

DROP TABLE IF EXISTS `problem_remote`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_remote` (
  `id` int NOT NULL AUTO_INCREMENT COMMENT '问题ID',
  `problem_id` int NOT NULL,
  `origin_oj` varchar(10) NOT NULL COMMENT '原始OJ标识',
  `origin_id` varchar(8) NOT NULL COMMENT '原始ID标识',
  `origin_url` varchar(100) NOT NULL COMMENT '原始链接',
  `origin_author` varchar(255) DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `problem_remote_pk` (`problem_id`),
  UNIQUE KEY `problem_origin` (`origin_oj`,`origin_id`)
) ENGINE=InnoDB AUTO_INCREMENT=333 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `problem_tag`
--

DROP TABLE IF EXISTS `problem_tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `problem_tag` (
  `id` int NOT NULL,
  `tag_id` int NOT NULL,
  `index` tinyint(1) NOT NULL,
  PRIMARY KEY (`id`,`tag_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `tag`
--

DROP TABLE IF EXISTS `tag`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `tag` (
  `id` int NOT NULL AUTO_INCREMENT,
  `name` varchar(20) CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci DEFAULT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `tag_pk` (`name`)
) ENGINE=InnoDB AUTO_INCREMENT=190 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user`
--

DROP TABLE IF EXISTS `user`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user` (
  `id` int NOT NULL AUTO_INCREMENT,
  `username` varchar(50) NOT NULL COMMENT '用户名',
  `nickname` varchar(80) NOT NULL COMMENT '昵称',
  `real_name` varchar(20) DEFAULT NULL COMMENT '真实名称',
  `password` char(80) NOT NULL,
  `email` varchar(90) DEFAULT NULL,
  `gender` tinyint(1) DEFAULT NULL COMMENT '性别',
  `number` varchar(20) DEFAULT NULL COMMENT '身份标识',
  `slogan` varchar(50) DEFAULT NULL COMMENT '签名',
  `organization` varchar(80) DEFAULT NULL COMMENT '组织',
  `qq` varchar(15) DEFAULT NULL COMMENT 'QQ',
  `vjudge_id` varchar(15) DEFAULT NULL COMMENT 'VjudgeId',
  `github` varchar(15) DEFAULT NULL COMMENT 'Github',
  `codeforces` varchar(20) DEFAULT NULL,
  `check_in_count` int NOT NULL COMMENT '签到次数',
  `insert_time` datetime NOT NULL,
  `modify_time` datetime NOT NULL,
  `accept` int NOT NULL,
  `attempt` int NOT NULL,
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_username` (`username`)
) ENGINE=InnoDB AUTO_INCREMENT=6623 DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;

--
-- Table structure for table `user_role`
--

DROP TABLE IF EXISTS `user_role`;
/*!40101 SET @saved_cs_client     = @@character_set_client */;
/*!50503 SET character_set_client = utf8mb4 */;
CREATE TABLE `user_role` (
  `id` int DEFAULT NULL,
  `role_id` varchar(10) DEFAULT NULL,
  UNIQUE KEY `group_id` (`id`,`role_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;
/*!40101 SET character_set_client = @saved_cs_client */;
/*!40103 SET TIME_ZONE=@OLD_TIME_ZONE */;

/*!40101 SET SQL_MODE=@OLD_SQL_MODE */;
/*!40014 SET FOREIGN_KEY_CHECKS=@OLD_FOREIGN_KEY_CHECKS */;
/*!40014 SET UNIQUE_CHECKS=@OLD_UNIQUE_CHECKS */;
/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
/*!40111 SET SQL_NOTES=@OLD_SQL_NOTES */;

-- Dump completed on 2025-10-20 18:25:18
