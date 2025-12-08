/*
 Navicat Premium Dump SQL

 Source Server         : aliyun
 Source Server Type    : PostgreSQL
 Source Server Version : 170006 (170006)
 Source Host           : pgm-2zeej00ms7k64wypko.pg.rds.aliyuncs.com:5432
 Source Catalog        : didaoj
 Source Schema         : didaoj

 Target Server Type    : PostgreSQL
 Target Server Version : 170006 (170006)
 File Encoding         : 65001

 Date: 08/12/2025 16:30:47
*/


-- ----------------------------
-- Sequence structure for bot_code_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."bot_code_id_seq";
CREATE SEQUENCE "didaoj"."bot_code_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for bot_game_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."bot_game_id_seq";
CREATE SEQUENCE "didaoj"."bot_game_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for bot_game_replay_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."bot_game_replay_id_seq";
CREATE SEQUENCE "didaoj"."bot_game_replay_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for collection_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."collection_id_seq";
CREATE SEQUENCE "didaoj"."collection_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for contest_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."contest_id_seq";
CREATE SEQUENCE "didaoj"."contest_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for discuss_comment_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."discuss_comment_id_seq";
CREATE SEQUENCE "didaoj"."discuss_comment_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for discuss_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."discuss_id_seq";
CREATE SEQUENCE "didaoj"."discuss_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for judge_job_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."judge_job_id_seq";
CREATE SEQUENCE "didaoj"."judge_job_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for problem_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."problem_id_seq";
CREATE SEQUENCE "didaoj"."problem_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for problem_local_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."problem_local_id_seq";
CREATE SEQUENCE "didaoj"."problem_local_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for problem_remote_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."problem_remote_id_seq";
CREATE SEQUENCE "didaoj"."problem_remote_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for run_job_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."run_job_id_seq";
CREATE SEQUENCE "didaoj"."run_job_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for tag_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."tag_id_seq";
CREATE SEQUENCE "didaoj"."tag_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for user_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."user_id_seq";
CREATE SEQUENCE "didaoj"."user_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for user_login_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."user_login_id_seq";
CREATE SEQUENCE "didaoj"."user_login_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Sequence structure for user_role_id_seq
-- ----------------------------
DROP SEQUENCE IF EXISTS "didaoj"."user_role_id_seq";
CREATE SEQUENCE "didaoj"."user_role_id_seq" 
INCREMENT 1
MINVALUE  1
MAXVALUE 9223372036854775807
START 1
CACHE 1;

-- ----------------------------
-- Table structure for bot_code
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."bot_code";
CREATE TABLE "didaoj"."bot_code" (
  "id" int4 NOT NULL GENERATED BY DEFAULT AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1
),
  "code" text COLLATE "pg_catalog"."default" NOT NULL,
  "version" int4 NOT NULL,
  "language" int2 NOT NULL,
  "inserter" int8 NOT NULL,
  "game_id" int4 NOT NULL,
  "insert_time" timestamptz(6),
  "modify_time" timestamptz(6)
)
;

-- ----------------------------
-- Table structure for bot_game
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."bot_game";
CREATE TABLE "didaoj"."bot_game" (
  "id" int4 NOT NULL GENERATED BY DEFAULT AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1
),
  "game_key" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default" NOT NULL,
  "title" varchar(30) COLLATE "pg_catalog"."default" NOT NULL,
  "judge_code" text COLLATE "pg_catalog"."default",
  "inserter" int4,
  "insert_time" timestamptz(6),
  "modify_time" timestamptz(6),
  "introduction" text COLLATE "pg_catalog"."default",
  "modifier" int4
)
;

-- ----------------------------
-- Table structure for bot_replay
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."bot_replay";
CREATE TABLE "didaoj"."bot_replay" (
  "id" int4 NOT NULL GENERATED BY DEFAULT AS IDENTITY (
INCREMENT 1
MINVALUE  1
MAXVALUE 2147483647
START 1
CACHE 1
),
  "param" text COLLATE "pg_catalog"."default" NOT NULL,
  "status" int2 NOT NULL,
  "info" text COLLATE "pg_catalog"."default",
  "game_id" int4 NOT NULL,
  "inserter" int4,
  "insert_time" timestamptz(6),
  "judger" varchar(10) COLLATE "pg_catalog"."default",
  "judge_time" timestamptz(6),
  "bots" int4[] NOT NULL,
  "message" text COLLATE "pg_catalog"."default"
)
;
COMMENT ON COLUMN "didaoj"."bot_replay"."param" IS '对局信息，一般用于记录可能会改变的信息，对局过程中客户端定期获取';
COMMENT ON COLUMN "didaoj"."bot_replay"."info" IS '初始信息，一般用于记录对局中不会改变的数据';

-- ----------------------------
-- Table structure for collection
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."collection";
CREATE TABLE "didaoj"."collection" (
  "id" int8 NOT NULL DEFAULT nextval('collection_id_seq'::regclass),
  "title" varchar(30) COLLATE "pg_catalog"."default",
  "description" text COLLATE "pg_catalog"."default",
  "start_time" timestamptz(6),
  "end_time" timestamptz(6),
  "inserter" int8,
  "insert_time" timestamptz(6),
  "modifier" int8,
  "modify_time" timestamptz(6),
  "private" bool,
  "password" varchar(30) COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for collection_member
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."collection_member";
CREATE TABLE "didaoj"."collection_member" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL
)
;

-- ----------------------------
-- Table structure for collection_problem
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."collection_problem";
CREATE TABLE "didaoj"."collection_problem" (
  "id" int8 NOT NULL,
  "problem_id" int8 NOT NULL,
  "index" int8
)
;

-- ----------------------------
-- Table structure for contest
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest";
CREATE TABLE "didaoj"."contest" (
  "id" int8 NOT NULL DEFAULT nextval('contest_id_seq'::regclass),
  "title" varchar(75) COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default",
  "notification" varchar(100) COLLATE "pg_catalog"."default",
  "start_time" timestamptz(6),
  "end_time" timestamptz(6),
  "inserter" int8,
  "insert_time" timestamptz(6),
  "modifier" int8,
  "modify_time" timestamptz(6),
  "private" bool,
  "password" varchar(35) COLLATE "pg_catalog"."default",
  "submit_anytime" bool,
  "type" int2,
  "score_type" int2,
  "lock_rank_duration" int8,
  "always_lock" bool,
  "discuss_type" int2,
  "notification_version" int4 NOT NULL
)
;

-- ----------------------------
-- Table structure for contest_language
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest_language";
CREATE TABLE "didaoj"."contest_language" (
  "id" int8 NOT NULL,
  "language" varchar(10) COLLATE "pg_catalog"."default" NOT NULL
)
;

-- ----------------------------
-- Table structure for contest_member
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest_member";
CREATE TABLE "didaoj"."contest_member" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL,
  "contest_name" varchar(20) COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for contest_member_auth
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest_member_auth";
CREATE TABLE "didaoj"."contest_member_auth" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL
)
;

-- ----------------------------
-- Table structure for contest_member_author
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest_member_author";
CREATE TABLE "didaoj"."contest_member_author" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL
)
;

-- ----------------------------
-- Table structure for contest_member_ignore
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest_member_ignore";
CREATE TABLE "didaoj"."contest_member_ignore" (
  "user_id" int8 NOT NULL,
  "id" int8 NOT NULL
)
;

-- ----------------------------
-- Table structure for contest_member_volunteer
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest_member_volunteer";
CREATE TABLE "didaoj"."contest_member_volunteer" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL
)
;

-- ----------------------------
-- Table structure for contest_problem
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest_problem";
CREATE TABLE "didaoj"."contest_problem" (
  "id" int8 NOT NULL,
  "problem_id" int8 NOT NULL,
  "index" int8,
  "view_id" int8,
  "score" int8
)
;

-- ----------------------------
-- Table structure for discuss
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."discuss";
CREATE TABLE "didaoj"."discuss" (
  "id" int8 NOT NULL DEFAULT nextval('discuss_id_seq'::regclass),
  "title" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "content" text COLLATE "pg_catalog"."default" NOT NULL,
  "view_count" int8 NOT NULL,
  "banned" bool,
  "problem_id" int8,
  "contest_id" int8,
  "judge_id" int8,
  "inserter" int8 NOT NULL,
  "insert_time" timestamptz(6) NOT NULL,
  "modifier" int8 NOT NULL,
  "modify_time" timestamptz(6) NOT NULL,
  "updater" int8 NOT NULL,
  "update_time" timestamptz(6) NOT NULL
)
;

-- ----------------------------
-- Table structure for discuss_comment
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."discuss_comment";
CREATE TABLE "didaoj"."discuss_comment" (
  "id" int8 NOT NULL DEFAULT nextval('discuss_comment_id_seq'::regclass),
  "discuss_id" int8 NOT NULL,
  "content" text COLLATE "pg_catalog"."default" NOT NULL,
  "banned" bool NOT NULL,
  "inserter" int8 NOT NULL,
  "insert_time" timestamptz(6) NOT NULL,
  "modifier" int8 NOT NULL,
  "modify_time" timestamptz(6) NOT NULL
)
;

-- ----------------------------
-- Table structure for discuss_tag
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."discuss_tag";
CREATE TABLE "didaoj"."discuss_tag" (
  "id" int8 NOT NULL,
  "tag_id" int8 NOT NULL,
  "index" int2 NOT NULL
)
;

-- ----------------------------
-- Table structure for judge_job
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."judge_job";
CREATE TABLE "didaoj"."judge_job" (
  "id" int8 NOT NULL DEFAULT nextval('judge_job_id_seq'::regclass),
  "problem_id" int8 NOT NULL,
  "contest_id" int8,
  "language" int2 NOT NULL,
  "code" text COLLATE "pg_catalog"."default" NOT NULL,
  "code_length" int8 NOT NULL,
  "status" int2 NOT NULL,
  "judger" varchar(10) COLLATE "pg_catalog"."default",
  "judge_time" timestamptz(6),
  "task_current" int8,
  "task_total" int8,
  "score" int8,
  "time" int8,
  "memory" int8,
  "private" bool NOT NULL,
  "remote_judge_id" varchar(20) COLLATE "pg_catalog"."default",
  "remote_account_id" varchar(20) COLLATE "pg_catalog"."default",
  "inserter" int8 NOT NULL,
  "insert_time" timestamptz(6) NOT NULL
)
;

-- ----------------------------
-- Table structure for judge_job_compile
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."judge_job_compile";
CREATE TABLE "didaoj"."judge_job_compile" (
  "id" int8 NOT NULL,
  "message" text COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for judge_task
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."judge_task";
CREATE TABLE "didaoj"."judge_task" (
  "id" int8 NOT NULL,
  "task_id" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "status" int2,
  "time" int8,
  "memory" int8,
  "score" int8,
  "content" text COLLATE "pg_catalog"."default",
  "hint" text COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for judger
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."judger";
CREATE TABLE "didaoj"."judger" (
  "key" varchar(10) COLLATE "pg_catalog"."default" NOT NULL,
  "name" varchar(20) COLLATE "pg_catalog"."default" NOT NULL,
  "max_job" int8,
  "cpu_usage" float8,
  "mem_usage" numeric,
  "mem_total" numeric,
  "avg_message" varchar(100) COLLATE "pg_catalog"."default",
  "insert_time" timestamptz(6),
  "modify_time" timestamptz(6),
  "hidden" bool,
  "enable" bool
)
;

-- ----------------------------
-- Table structure for kv_store
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."kv_store";
CREATE TABLE "didaoj"."kv_store" (
  "key" text COLLATE "pg_catalog"."default" NOT NULL,
  "value" text COLLATE "pg_catalog"."default",
  "insert_time" timestamptz(6) DEFAULT now(),
  "expire_time" timestamptz(6)
)
;

-- ----------------------------
-- Table structure for problem
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."problem";
CREATE TABLE "didaoj"."problem" (
  "id" int8 NOT NULL DEFAULT nextval('problem_id_seq'::regclass),
  "key" varchar(15) COLLATE "pg_catalog"."default",
  "title" varchar(50) COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default" NOT NULL,
  "source" varchar(300) COLLATE "pg_catalog"."default",
  "time_limit" int8 NOT NULL,
  "memory_limit" int8 NOT NULL,
  "judge_type" int2 NOT NULL,
  "inserter" int8 NOT NULL,
  "insert_time" timestamptz(6) NOT NULL,
  "modifier" int8 NOT NULL,
  "modify_time" timestamptz(6) NOT NULL,
  "accept" int8 NOT NULL,
  "attempt" int8 NOT NULL,
  "private" bool NOT NULL
)
;

-- ----------------------------
-- Table structure for problem_daily
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."problem_daily";
CREATE TABLE "didaoj"."problem_daily" (
  "key" char(10) COLLATE "pg_catalog"."default" NOT NULL,
  "problem_id" int8 NOT NULL,
  "solution" text COLLATE "pg_catalog"."default" NOT NULL,
  "code" text COLLATE "pg_catalog"."default" NOT NULL,
  "inserter" int8,
  "modifier" int8,
  "insert_time" timestamptz(6),
  "modify_time" timestamptz(6)
)
;

-- ----------------------------
-- Table structure for problem_local
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."problem_local";
CREATE TABLE "didaoj"."problem_local" (
  "id" int8 NOT NULL DEFAULT nextval('problem_local_id_seq'::regclass),
  "problem_id" int8 NOT NULL,
  "judge_md5" char(32) COLLATE "pg_catalog"."default",
  "judge_job" jsonb
)
;

-- ----------------------------
-- Table structure for problem_member
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."problem_member";
CREATE TABLE "didaoj"."problem_member" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL
)
;

-- ----------------------------
-- Table structure for problem_member_auth
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."problem_member_auth";
CREATE TABLE "didaoj"."problem_member_auth" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL
)
;

-- ----------------------------
-- Table structure for problem_remote
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."problem_remote";
CREATE TABLE "didaoj"."problem_remote" (
  "id" int8 NOT NULL DEFAULT nextval('problem_remote_id_seq'::regclass),
  "problem_id" int8 NOT NULL,
  "origin_oj" varchar(10) COLLATE "pg_catalog"."default" NOT NULL,
  "origin_id" varchar(8) COLLATE "pg_catalog"."default" NOT NULL,
  "origin_url" varchar(100) COLLATE "pg_catalog"."default" NOT NULL,
  "origin_author" varchar(255) COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for problem_tag
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."problem_tag";
CREATE TABLE "didaoj"."problem_tag" (
  "id" int8 NOT NULL,
  "tag_id" int8 NOT NULL,
  "index" int2 NOT NULL
)
;

-- ----------------------------
-- Table structure for run_job
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."run_job";
CREATE TABLE "didaoj"."run_job" (
  "id" int8 NOT NULL DEFAULT nextval('run_job_id_seq'::regclass),
  "inserter" int8 NOT NULL,
  "code" text COLLATE "pg_catalog"."default" NOT NULL,
  "input" text COLLATE "pg_catalog"."default",
  "language" int2 NOT NULL,
  "content" text COLLATE "pg_catalog"."default",
  "status" int2 NOT NULL DEFAULT 0,
  "time" int8 DEFAULT 0,
  "memory" int8 DEFAULT 0,
  "insert_time" timestamptz(6) NOT NULL DEFAULT CURRENT_TIMESTAMP,
  "judger" varchar(10) COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for tag
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."tag";
CREATE TABLE "didaoj"."tag" (
  "id" int8 NOT NULL DEFAULT nextval('tag_id_seq'::regclass),
  "name" varchar(20) COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for user
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."user";
CREATE TABLE "didaoj"."user" (
  "id" int8 NOT NULL DEFAULT nextval('user_id_seq'::regclass),
  "username" varchar(50) COLLATE "pg_catalog"."default" NOT NULL,
  "nickname" varchar(80) COLLATE "pg_catalog"."default" NOT NULL,
  "real_name" varchar(20) COLLATE "pg_catalog"."default",
  "password" varchar(80) COLLATE "pg_catalog"."default" NOT NULL,
  "email" varchar(90) COLLATE "pg_catalog"."default",
  "gender" int2,
  "number" varchar(20) COLLATE "pg_catalog"."default",
  "slogan" varchar(50) COLLATE "pg_catalog"."default",
  "organization" varchar(80) COLLATE "pg_catalog"."default",
  "qq" varchar(15) COLLATE "pg_catalog"."default",
  "vjudge_id" varchar(15) COLLATE "pg_catalog"."default",
  "github" varchar(15) COLLATE "pg_catalog"."default",
  "codeforces" varchar(20) COLLATE "pg_catalog"."default",
  "check_in_count" int8 NOT NULL,
  "insert_time" timestamptz(6) NOT NULL,
  "modify_time" timestamptz(6) NOT NULL,
  "accept" int8 NOT NULL,
  "attempt" int8 NOT NULL,
  "blog" varchar(100) COLLATE "pg_catalog"."default",
  "level" int4,
  "experience" int4,
  "coin" int4 NOT NULL DEFAULT 0
)
;

-- ----------------------------
-- Table structure for user_coin
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."user_coin";
CREATE TABLE "didaoj"."user_coin" (
  "user_id" int8,
  "value" int4,
  "inserter_time" timestamp(6),
  "type" int4,
  "param" varchar(255) COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for user_experience
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."user_experience";
CREATE TABLE "didaoj"."user_experience" (
  "user_id" int8,
  "value" int4,
  "inserter_time" timestamp(6),
  "type" int4,
  "param" varchar(255) COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for user_login
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."user_login";
CREATE TABLE "didaoj"."user_login" (
  "id" int8 NOT NULL DEFAULT nextval('user_login_id_seq'::regclass),
  "user_id" int4 NOT NULL,
  "insert_time" timestamptz(6) NOT NULL,
  "ip" inet NOT NULL,
  "user_agent" text COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for user_role
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."user_role";
CREATE TABLE "didaoj"."user_role" (
  "id" int8 NOT NULL DEFAULT nextval('user_role_id_seq'::regclass),
  "role_id" varchar(10) COLLATE "pg_catalog"."default" NOT NULL
)
;

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."bot_code_id_seq"
OWNED BY "didaoj"."bot_code"."id";
SELECT setval('"didaoj"."bot_code_id_seq"', 1, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."bot_game_id_seq"
OWNED BY "didaoj"."bot_game"."id";
SELECT setval('"didaoj"."bot_game_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."bot_game_replay_id_seq"
OWNED BY "didaoj"."bot_replay"."id";
SELECT setval('"didaoj"."bot_game_replay_id_seq"', 1, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."collection_id_seq"
OWNED BY "didaoj"."collection"."id";
SELECT setval('"didaoj"."collection_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."contest_id_seq"
OWNED BY "didaoj"."contest"."id";
SELECT setval('"didaoj"."contest_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."discuss_comment_id_seq"
OWNED BY "didaoj"."discuss_comment"."id";
SELECT setval('"didaoj"."discuss_comment_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."discuss_id_seq"
OWNED BY "didaoj"."discuss"."id";
SELECT setval('"didaoj"."discuss_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."judge_job_id_seq"
OWNED BY "didaoj"."judge_job"."id";
SELECT setval('"didaoj"."judge_job_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."problem_id_seq"
OWNED BY "didaoj"."problem"."id";
SELECT setval('"didaoj"."problem_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."problem_local_id_seq"
OWNED BY "didaoj"."problem_local"."id";
SELECT setval('"didaoj"."problem_local_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."problem_remote_id_seq"
OWNED BY "didaoj"."problem_remote"."id";
SELECT setval('"didaoj"."problem_remote_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."run_job_id_seq"
OWNED BY "didaoj"."run_job"."id";
SELECT setval('"didaoj"."run_job_id_seq"', 1, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."tag_id_seq"
OWNED BY "didaoj"."tag"."id";
SELECT setval('"didaoj"."tag_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."user_id_seq"
OWNED BY "didaoj"."user"."id";
SELECT setval('"didaoj"."user_id_seq"', 1, false);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
ALTER SEQUENCE "didaoj"."user_login_id_seq"
OWNED BY "didaoj"."user_login"."id";
SELECT setval('"didaoj"."user_login_id_seq"', 1, true);

-- ----------------------------
-- Alter sequences owned by
-- ----------------------------
SELECT setval('"didaoj"."user_role_id_seq"', 1, true);

-- ----------------------------
-- Indexes structure for table bot_code
-- ----------------------------
CREATE INDEX "bot_code_game_id_index" ON "didaoj"."bot_code" USING btree (
  "game_id" "pg_catalog"."int4_ops" ASC NULLS LAST
);
CREATE INDEX "bot_code_inserter_index" ON "didaoj"."bot_code" USING btree (
  "inserter" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "bot_code_language_index" ON "didaoj"."bot_code" USING btree (
  "language" "pg_catalog"."int2_ops" ASC NULLS LAST
);
CREATE INDEX "bot_code_version_index" ON "didaoj"."bot_code" USING btree (
  "version" "pg_catalog"."int4_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table bot_code
-- ----------------------------
ALTER TABLE "didaoj"."bot_code" ADD CONSTRAINT "bot_code_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table bot_game
-- ----------------------------
CREATE INDEX "bot_game_game_key_index" ON "didaoj"."bot_game" USING btree (
  "game_key" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Uniques structure for table bot_game
-- ----------------------------
ALTER TABLE "didaoj"."bot_game" ADD CONSTRAINT "bot_game_pk_2" UNIQUE ("game_key");

-- ----------------------------
-- Primary Key structure for table bot_game
-- ----------------------------
ALTER TABLE "didaoj"."bot_game" ADD CONSTRAINT "bot_game_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table bot_replay
-- ----------------------------
CREATE INDEX "bot_replay_status_index" ON "didaoj"."bot_replay" USING btree (
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table bot_replay
-- ----------------------------
ALTER TABLE "didaoj"."bot_replay" ADD CONSTRAINT "bot_replay_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table collection
-- ----------------------------
ALTER TABLE "didaoj"."collection" ADD CONSTRAINT "collection_id_idx" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table collection_member
-- ----------------------------
ALTER TABLE "didaoj"."collection_member" ADD CONSTRAINT "collection_member_pk" PRIMARY KEY ("id", "user_id");

-- ----------------------------
-- Primary Key structure for table collection_problem
-- ----------------------------
ALTER TABLE "didaoj"."collection_problem" ADD CONSTRAINT "collection_problem_pk" PRIMARY KEY ("id", "problem_id");

-- ----------------------------
-- Primary Key structure for table contest
-- ----------------------------
ALTER TABLE "didaoj"."contest" ADD CONSTRAINT "contest_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table contest_language
-- ----------------------------
ALTER TABLE "didaoj"."contest_language" ADD CONSTRAINT "contest_language_pk" PRIMARY KEY ("language", "id");

-- ----------------------------
-- Primary Key structure for table contest_member
-- ----------------------------
ALTER TABLE "didaoj"."contest_member" ADD CONSTRAINT "contest_member_pk" PRIMARY KEY ("user_id", "id");

-- ----------------------------
-- Primary Key structure for table contest_member_auth
-- ----------------------------
ALTER TABLE "didaoj"."contest_member_auth" ADD CONSTRAINT "contest_member_auth_pk" PRIMARY KEY ("user_id", "id");

-- ----------------------------
-- Primary Key structure for table contest_member_author
-- ----------------------------
ALTER TABLE "didaoj"."contest_member_author" ADD CONSTRAINT "contest_member_author_pk" PRIMARY KEY ("user_id", "id");

-- ----------------------------
-- Primary Key structure for table contest_member_ignore
-- ----------------------------
ALTER TABLE "didaoj"."contest_member_ignore" ADD CONSTRAINT "contest_member_ignore_pk" PRIMARY KEY ("id", "user_id");

-- ----------------------------
-- Primary Key structure for table contest_member_volunteer
-- ----------------------------
ALTER TABLE "didaoj"."contest_member_volunteer" ADD CONSTRAINT "contest_member_volunteer_pk" PRIMARY KEY ("user_id", "id");

-- ----------------------------
-- Primary Key structure for table contest_problem
-- ----------------------------
ALTER TABLE "didaoj"."contest_problem" ADD CONSTRAINT "contest_problem_pk" PRIMARY KEY ("problem_id", "id");

-- ----------------------------
-- Indexes structure for table discuss
-- ----------------------------
CREATE INDEX "discuss_contest_id_index" ON "didaoj"."discuss" USING btree (
  "contest_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "discuss_inserter_index" ON "didaoj"."discuss" USING btree (
  "inserter" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "discuss_judge_id_index" ON "didaoj"."discuss" USING btree (
  "judge_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "discuss_problem_id_index" ON "didaoj"."discuss" USING btree (
  "problem_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table discuss
-- ----------------------------
ALTER TABLE "didaoj"."discuss" ADD CONSTRAINT "discuss_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table discuss_comment
-- ----------------------------
CREATE INDEX "discuss_comment_discuss_id_index" ON "didaoj"."discuss_comment" USING btree (
  "discuss_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table discuss_comment
-- ----------------------------
ALTER TABLE "didaoj"."discuss_comment" ADD CONSTRAINT "discuss_comment_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table discuss_tag
-- ----------------------------
ALTER TABLE "didaoj"."discuss_tag" ADD CONSTRAINT "discuss_tag_pk" PRIMARY KEY ("id", "tag_id");

-- ----------------------------
-- Indexes structure for table judge_job
-- ----------------------------
CREATE INDEX "judge_job_contest_id_index" ON "didaoj"."judge_job" USING btree (
  "contest_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "judge_job_inserter_index" ON "didaoj"."judge_job" USING btree (
  "inserter" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "judge_job_judger_index" ON "didaoj"."judge_job" USING btree (
  "judger" COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "judge_job_language_index" ON "didaoj"."judge_job" USING btree (
  "language" "pg_catalog"."int2_ops" ASC NULLS LAST
);
CREATE INDEX "judge_job_problem_id_index" ON "didaoj"."judge_job" USING btree (
  "problem_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "judge_job_status_index" ON "didaoj"."judge_job" USING btree (
  "status" "pg_catalog"."int2_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table judge_job
-- ----------------------------
ALTER TABLE "didaoj"."judge_job" ADD CONSTRAINT "judge_job_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table judge_job_compile
-- ----------------------------
ALTER TABLE "didaoj"."judge_job_compile" ADD CONSTRAINT "judge_job_compile_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table judge_task
-- ----------------------------
ALTER TABLE "didaoj"."judge_task" ADD CONSTRAINT "judge_task_pk" PRIMARY KEY ("task_id", "id");

-- ----------------------------
-- Primary Key structure for table judger
-- ----------------------------
ALTER TABLE "didaoj"."judger" ADD CONSTRAINT "judger_pk" PRIMARY KEY ("key");

-- ----------------------------
-- Indexes structure for table kv_store
-- ----------------------------
CREATE INDEX "kv_store_expire_time_index" ON "didaoj"."kv_store" USING btree (
  "expire_time" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "kv_store_insert_time_index" ON "didaoj"."kv_store" USING btree (
  "insert_time" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table kv_store
-- ----------------------------
ALTER TABLE "didaoj"."kv_store" ADD CONSTRAINT "kv_store_pkey" PRIMARY KEY ("key");

-- ----------------------------
-- Indexes structure for table problem
-- ----------------------------
CREATE INDEX "idx_problem_key_lower" ON "didaoj"."problem" USING btree (
  lower(key::text) COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_problem_title_trgm" ON "didaoj"."problem" USING gin (
  "title" COLLATE "pg_catalog"."default" "public"."gin_trgm_ops"
);
CREATE INDEX "problem_inserter_index" ON "didaoj"."problem" USING btree (
  "inserter" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Uniques structure for table problem
-- ----------------------------
ALTER TABLE "didaoj"."problem" ADD CONSTRAINT "problem_pk_2" UNIQUE ("key");

-- ----------------------------
-- Primary Key structure for table problem
-- ----------------------------
ALTER TABLE "didaoj"."problem" ADD CONSTRAINT "problem_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table problem_daily
-- ----------------------------
CREATE INDEX "problem_daily_inserter_index" ON "didaoj"."problem_daily" USING btree (
  "inserter" "pg_catalog"."int8_ops" ASC NULLS LAST
);
CREATE INDEX "problem_daily_problem_id_index" ON "didaoj"."problem_daily" USING btree (
  "problem_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table problem_daily
-- ----------------------------
ALTER TABLE "didaoj"."problem_daily" ADD CONSTRAINT "problem_daily_pk" PRIMARY KEY ("key");

-- ----------------------------
-- Indexes structure for table problem_local
-- ----------------------------
CREATE INDEX "problem_local_problem_id_index" ON "didaoj"."problem_local" USING btree (
  "problem_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table problem_local
-- ----------------------------
ALTER TABLE "didaoj"."problem_local" ADD CONSTRAINT "problem_local_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table problem_member
-- ----------------------------
ALTER TABLE "didaoj"."problem_member" ADD CONSTRAINT "problem_member_pk" PRIMARY KEY ("user_id", "id");

-- ----------------------------
-- Primary Key structure for table problem_member_auth
-- ----------------------------
ALTER TABLE "didaoj"."problem_member_auth" ADD CONSTRAINT "problem_member_auth_pk" PRIMARY KEY ("id", "user_id");

-- ----------------------------
-- Indexes structure for table problem_remote
-- ----------------------------
CREATE INDEX "problem_remote_problem_id_index" ON "didaoj"."problem_remote" USING btree (
  "problem_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table problem_remote
-- ----------------------------
ALTER TABLE "didaoj"."problem_remote" ADD CONSTRAINT "problem_remote_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table problem_tag
-- ----------------------------
ALTER TABLE "didaoj"."problem_tag" ADD CONSTRAINT "problem_tag_pk" PRIMARY KEY ("id", "tag_id");

-- ----------------------------
-- Indexes structure for table run_job
-- ----------------------------
CREATE INDEX "idx_run_job_insert_time" ON "didaoj"."run_job" USING btree (
  "insert_time" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "idx_run_job_user_id" ON "didaoj"."run_job" USING btree (
  "inserter" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table run_job
-- ----------------------------
ALTER TABLE "didaoj"."run_job" ADD CONSTRAINT "run_job_pkey" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table tag
-- ----------------------------
CREATE INDEX "idx_tag_name_trgm" ON "didaoj"."tag" USING gin (
  "name" COLLATE "pg_catalog"."default" "public"."gin_trgm_ops"
);

-- ----------------------------
-- Uniques structure for table tag
-- ----------------------------
ALTER TABLE "didaoj"."tag" ADD CONSTRAINT "tag_pk_2" UNIQUE ("name");

-- ----------------------------
-- Primary Key structure for table tag
-- ----------------------------
ALTER TABLE "didaoj"."tag" ADD CONSTRAINT "tag_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table user
-- ----------------------------
CREATE UNIQUE INDEX "idx_user_username_lower" ON "didaoj"."user" USING btree (
  lower(username::text) COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);

-- ----------------------------
-- Uniques structure for table user
-- ----------------------------
ALTER TABLE "didaoj"."user" ADD CONSTRAINT "user_pk_2" UNIQUE ("username");

-- ----------------------------
-- Primary Key structure for table user
-- ----------------------------
ALTER TABLE "didaoj"."user" ADD CONSTRAINT "user_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Indexes structure for table user_coin
-- ----------------------------
CREATE INDEX "user_coin_user_id_index" ON "didaoj"."user_coin" USING btree (
  "user_id" "pg_catalog"."int8_ops" ASC NULLS LAST
);

-- ----------------------------
-- Uniques structure for table user_experience
-- ----------------------------
ALTER TABLE "didaoj"."user_experience" ADD CONSTRAINT "user_experience_pk" UNIQUE ("type", "user_id", "param");

-- ----------------------------
-- Indexes structure for table user_login
-- ----------------------------
CREATE INDEX "user_login_insert_time_index" ON "didaoj"."user_login" USING btree (
  "insert_time" "pg_catalog"."timestamptz_ops" ASC NULLS LAST
);
CREATE INDEX "user_login_user_id_index" ON "didaoj"."user_login" USING btree (
  "user_id" "pg_catalog"."int4_ops" ASC NULLS LAST
);

-- ----------------------------
-- Primary Key structure for table user_login
-- ----------------------------
ALTER TABLE "didaoj"."user_login" ADD CONSTRAINT "user_login_pk" PRIMARY KEY ("id");

-- ----------------------------
-- Primary Key structure for table user_role
-- ----------------------------
ALTER TABLE "didaoj"."user_role" ADD CONSTRAINT "user_role_pk" PRIMARY KEY ("role_id", "id");

-- ----------------------------
-- Foreign Keys structure for table bot_code
-- ----------------------------
ALTER TABLE "didaoj"."bot_code" ADD CONSTRAINT "bot_code_bot_game_id_fk" FOREIGN KEY ("game_id") REFERENCES "didaoj"."bot_game" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
ALTER TABLE "didaoj"."bot_code" ADD CONSTRAINT "bot_code_user_id_fk" FOREIGN KEY ("inserter") REFERENCES "didaoj"."user" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;

-- ----------------------------
-- Foreign Keys structure for table user_experience
-- ----------------------------
ALTER TABLE "didaoj"."user_experience" ADD CONSTRAINT "user_experience_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "didaoj"."user" ("id") ON DELETE NO ACTION ON UPDATE NO ACTION;
