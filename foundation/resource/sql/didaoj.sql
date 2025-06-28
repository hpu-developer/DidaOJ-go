-- phpMyAdmin SQL Dump
-- version 5.2.2
-- https://www.phpmyadmin.net/
--
-- 主机： 1Panel-mysql-6wqo:3306
-- 生成日期： 2025-06-28 09:34:20
-- 服务器版本： 8.4.5
-- PHP 版本： 8.2.28

SET SQL_MODE = "NO_AUTO_VALUE_ON_ZERO";
START TRANSACTION;
SET time_zone = "+00:00";


/*!40101 SET @OLD_CHARACTER_SET_CLIENT=@@CHARACTER_SET_CLIENT */;
/*!40101 SET @OLD_CHARACTER_SET_RESULTS=@@CHARACTER_SET_RESULTS */;
/*!40101 SET @OLD_COLLATION_CONNECTION=@@COLLATION_CONNECTION */;
/*!40101 SET NAMES utf8mb4 */;

--
-- 数据库： `didaoj`
--

-- --------------------------------------------------------

--
-- 表的结构 `collection`
--

CREATE TABLE `collection` (
  `id` int NOT NULL,
  `title` varchar(30) DEFAULT NULL,
  `description` text,
  `start_time` datetime DEFAULT NULL,
  `end_time` datetime DEFAULT NULL,
  `inserter` int DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modifier` int DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL,
  `private` tinyint(1) DEFAULT NULL,
  `password` varchar(30) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `collection_member`
--

CREATE TABLE `collection_member` (
  `id` int NOT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='题集参与人员';

-- --------------------------------------------------------

--
-- 表的结构 `collection_problem`
--

CREATE TABLE `collection_problem` (
  `id` int NOT NULL,
  `problem_id` int NOT NULL,
  `index` int DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛限制语言';

-- --------------------------------------------------------

--
-- 表的结构 `contest`
--

CREATE TABLE `contest` (
  `id` int NOT NULL,
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
  `discuss_type` tinyint DEFAULT NULL COMMENT '讨论类型，0正常讨论，1仅查看自己的讨论'
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `contest_language`
--

CREATE TABLE `contest_language` (
  `id` int NOT NULL,
  `language` varchar(10) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛限制语言';

-- --------------------------------------------------------

--
-- 表的结构 `contest_member`
--

CREATE TABLE `contest_member` (
  `id` int NOT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛访问权限';

-- --------------------------------------------------------

--
-- 表的结构 `contest_member_auth`
--

CREATE TABLE `contest_member_auth` (
  `id` int NOT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛管理权限';

-- --------------------------------------------------------

--
-- 表的结构 `contest_member_author`
--

CREATE TABLE `contest_member_author` (
  `id` int NOT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛作者列表';

-- --------------------------------------------------------

--
-- 表的结构 `contest_member_ignore`
--

CREATE TABLE `contest_member_ignore` (
  `user_id` int NOT NULL,
  `id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛忽略排行成员';

-- --------------------------------------------------------

--
-- 表的结构 `contest_member_volunteer`
--

CREATE TABLE `contest_member_volunteer` (
  `id` int NOT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛志愿者权限';

-- --------------------------------------------------------

--
-- 表的结构 `contest_problem`
--

CREATE TABLE `contest_problem` (
  `id` int NOT NULL,
  `problem_id` int DEFAULT NULL,
  `index` int DEFAULT NULL,
  `view_id` int DEFAULT NULL,
  `score` int DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='比赛限制语言';

-- --------------------------------------------------------

--
-- 表的结构 `discuss`
--

CREATE TABLE `discuss` (
  `id` int NOT NULL,
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
  `update_time` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `discuss_comment`
--

CREATE TABLE `discuss_comment` (
  `id` int NOT NULL,
  `discuss_id` int NOT NULL,
  `content` text NOT NULL,
  `banned` tinyint(1) NOT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL,
  `modifier` int NOT NULL,
  `modify_time` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `discuss_tag`
--

CREATE TABLE `discuss_tag` (
  `id` int NOT NULL,
  `tag_id` int NOT NULL,
  `index` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `judger`
--

CREATE TABLE `judger` (
  `key` varchar(10) NOT NULL,
  `name` varchar(20) NOT NULL,
  `max_job` int DEFAULT NULL,
  `cpu_usage` float DEFAULT NULL,
  `mem_usage` bigint UNSIGNED DEFAULT NULL,
  `mem_total` bigint UNSIGNED DEFAULT NULL,
  `avg_message` varchar(100) DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `judge_job`
--

CREATE TABLE `judge_job` (
  `id` int NOT NULL,
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
  `time` int DEFAULT NULL,
  `memory` int DEFAULT NULL,
  `private` tinyint(1) NOT NULL,
  `remote_judge_id` varchar(20) DEFAULT NULL,
  `remote_account_id` varchar(20) DEFAULT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `judge_job_compile`
--

CREATE TABLE `judge_job_compile` (
  `id` int DEFAULT NULL,
  `message` text
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `problem`
--

CREATE TABLE `problem` (
  `id` int NOT NULL COMMENT '问题ID',
  `key` varchar(18) NOT NULL COMMENT '问题对外标识',
  `title` varchar(50) NOT NULL COMMENT '标题',
  `description` text NOT NULL COMMENT '描述',
  `source` varchar(120) DEFAULT NULL,
  `time_limit` int NOT NULL COMMENT '时间限制',
  `memory_limit` int NOT NULL COMMENT '内存限制',
  `judge_type` tinyint(1) NOT NULL,
  `inserter` int NOT NULL,
  `insert_time` datetime NOT NULL COMMENT '创建时间',
  `modifier` int NOT NULL,
  `modify_time` datetime NOT NULL COMMENT '创建时间',
  `accept` int NOT NULL COMMENT '通过人数',
  `attempt` int NOT NULL COMMENT '尝试人数',
  `private` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `problem_daily`
--

CREATE TABLE `problem_daily` (
  `key` char(10) NOT NULL,
  `problem_id` int NOT NULL,
  `solution` text NOT NULL,
  `code` text NOT NULL,
  `inserter` int DEFAULT NULL,
  `modifier` int DEFAULT NULL,
  `insert_time` datetime DEFAULT NULL,
  `modify_time` datetime DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `problem_local`
--

CREATE TABLE `problem_local` (
  `id` int NOT NULL COMMENT '问题ID',
  `problem_id` int NOT NULL,
  `judge_md5` char(32) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `problem_member`
--

CREATE TABLE `problem_member` (
  `id` int NOT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='问题访问权限';

-- --------------------------------------------------------

--
-- 表的结构 `problem_member_auth`
--

CREATE TABLE `problem_member_auth` (
  `id` int NOT NULL,
  `user_id` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci COMMENT='问题管理权限';

-- --------------------------------------------------------

--
-- 表的结构 `problem_remote`
--

CREATE TABLE `problem_remote` (
  `id` int NOT NULL COMMENT '问题ID',
  `problem_id` int NOT NULL,
  `origin_oj` varchar(10) NOT NULL COMMENT '原始OJ标识',
  `origin_id` varchar(8) NOT NULL COMMENT '原始ID标识',
  `origin_url` varchar(100) NOT NULL COMMENT '原始链接',
  `origin_author` varchar(255) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `problem_tag`
--

CREATE TABLE `problem_tag` (
  `id` int NOT NULL,
  `tag_id` int NOT NULL,
  `index` tinyint(1) NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `tag`
--

CREATE TABLE `tag` (
  `id` int NOT NULL,
  `name` varchar(20) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `user`
--

CREATE TABLE `user` (
  `id` int NOT NULL,
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
  `attempt` int NOT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

-- --------------------------------------------------------

--
-- 表的结构 `user_role`
--

CREATE TABLE `user_role` (
  `id` int DEFAULT NULL,
  `role_id` varchar(10) DEFAULT NULL
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

--
-- 转储表的索引
--

--
-- 表的索引 `collection`
--
ALTER TABLE `collection`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `collection_member`
--
ALTER TABLE `collection_member`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `collection_problem`
--
ALTER TABLE `collection_problem`
  ADD PRIMARY KEY (`id`,`problem_id`);

--
-- 表的索引 `contest`
--
ALTER TABLE `contest`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `contest_language`
--
ALTER TABLE `contest_language`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `contest_member`
--
ALTER TABLE `contest_member`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `contest_member_auth`
--
ALTER TABLE `contest_member_auth`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `contest_member_author`
--
ALTER TABLE `contest_member_author`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `contest_member_ignore`
--
ALTER TABLE `contest_member_ignore`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `contest_member_volunteer`
--
ALTER TABLE `contest_member_volunteer`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `contest_problem`
--
ALTER TABLE `contest_problem`
  ADD UNIQUE KEY `group_id` (`id`,`problem_id`),
  ADD UNIQUE KEY `group_index` (`id`,`index`);

--
-- 表的索引 `discuss`
--
ALTER TABLE `discuss`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `discuss_pk_2` (`id`);

--
-- 表的索引 `discuss_comment`
--
ALTER TABLE `discuss_comment`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `discuss_pk` (`id`);

--
-- 表的索引 `discuss_tag`
--
ALTER TABLE `discuss_tag`
  ADD PRIMARY KEY (`id`,`tag_id`);

--
-- 表的索引 `judger`
--
ALTER TABLE `judger`
  ADD PRIMARY KEY (`key`);

--
-- 表的索引 `judge_job`
--
ALTER TABLE `judge_job`
  ADD PRIMARY KEY (`id`);

--
-- 表的索引 `judge_job_compile`
--
ALTER TABLE `judge_job_compile`
  ADD UNIQUE KEY `judge_job_compile_pk` (`id`);

--
-- 表的索引 `problem`
--
ALTER TABLE `problem`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `problem_pk` (`key`);

--
-- 表的索引 `problem_daily`
--
ALTER TABLE `problem_daily`
  ADD PRIMARY KEY (`key`),
  ADD UNIQUE KEY `problem_daily_pk_2` (`problem_id`);

--
-- 表的索引 `problem_local`
--
ALTER TABLE `problem_local`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `problem_local_pk` (`problem_id`);

--
-- 表的索引 `problem_member`
--
ALTER TABLE `problem_member`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `problem_member_auth`
--
ALTER TABLE `problem_member_auth`
  ADD PRIMARY KEY (`id`,`user_id`);

--
-- 表的索引 `problem_remote`
--
ALTER TABLE `problem_remote`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `problem_remote_pk` (`problem_id`),
  ADD UNIQUE KEY `problem_origin` (`origin_oj`,`origin_id`);

--
-- 表的索引 `problem_tag`
--
ALTER TABLE `problem_tag`
  ADD PRIMARY KEY (`id`,`tag_id`);

--
-- 表的索引 `tag`
--
ALTER TABLE `tag`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `tag_pk` (`name`);

--
-- 表的索引 `user`
--
ALTER TABLE `user`
  ADD PRIMARY KEY (`id`),
  ADD UNIQUE KEY `uk_username` (`username`);

--
-- 表的索引 `user_role`
--
ALTER TABLE `user_role`
  ADD UNIQUE KEY `group_id` (`id`,`role_id`);

--
-- 在导出的表使用AUTO_INCREMENT
--

--
-- 使用表AUTO_INCREMENT `collection`
--
ALTER TABLE `collection`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `contest`
--
ALTER TABLE `contest`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `discuss`
--
ALTER TABLE `discuss`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `discuss_comment`
--
ALTER TABLE `discuss_comment`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `judge_job`
--
ALTER TABLE `judge_job`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `problem`
--
ALTER TABLE `problem`
  MODIFY `id` int NOT NULL AUTO_INCREMENT COMMENT '问题ID';

--
-- 使用表AUTO_INCREMENT `problem_local`
--
ALTER TABLE `problem_local`
  MODIFY `id` int NOT NULL AUTO_INCREMENT COMMENT '问题ID';

--
-- 使用表AUTO_INCREMENT `problem_remote`
--
ALTER TABLE `problem_remote`
  MODIFY `id` int NOT NULL AUTO_INCREMENT COMMENT '问题ID';

--
-- 使用表AUTO_INCREMENT `tag`
--
ALTER TABLE `tag`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;

--
-- 使用表AUTO_INCREMENT `user`
--
ALTER TABLE `user`
  MODIFY `id` int NOT NULL AUTO_INCREMENT;
COMMIT;

/*!40101 SET CHARACTER_SET_CLIENT=@OLD_CHARACTER_SET_CLIENT */;
/*!40101 SET CHARACTER_SET_RESULTS=@OLD_CHARACTER_SET_RESULTS */;
/*!40101 SET COLLATION_CONNECTION=@OLD_COLLATION_CONNECTION */;
