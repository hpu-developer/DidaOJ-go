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

 Date: 27/10/2025 22:53:03
*/


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
  "discuss_type" int2
)
;

-- ----------------------------
-- Table structure for contest_language
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest_language";
CREATE TABLE "didaoj"."contest_language" (
  "id" int8 NOT NULL,
  "language" varchar(10) COLLATE "pg_catalog"."default"
)
;

-- ----------------------------
-- Table structure for contest_member
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."contest_member";
CREATE TABLE "didaoj"."contest_member" (
  "id" int8 NOT NULL,
  "user_id" int8 NOT NULL
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
  "problem_id" int8,
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
  "id" int8,
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
-- Table structure for problem
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."problem";
CREATE TABLE "didaoj"."problem" (
  "id" int8 NOT NULL DEFAULT nextval('problem_id_seq'::regclass),
  "key" varchar(15) COLLATE "pg_catalog"."default",
  "title" varchar(50) COLLATE "pg_catalog"."default" NOT NULL,
  "description" text COLLATE "pg_catalog"."default" NOT NULL,
  "source" varchar(250) COLLATE "pg_catalog"."default",
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
  "judge_md5" char(32) COLLATE "pg_catalog"."default"
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
  "attempt" int8 NOT NULL
)
;

-- ----------------------------
-- Table structure for user_role
-- ----------------------------
DROP TABLE IF EXISTS "didaoj"."user_role";
CREATE TABLE "didaoj"."user_role" (
  "id" int8,
  "role_id" varchar(10) COLLATE "pg_catalog"."default"
)
;

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
-- Uniques structure for table judger
-- ----------------------------
ALTER TABLE "didaoj"."judger" ADD CONSTRAINT "judger_key_unique" UNIQUE ("key");

-- ----------------------------
-- Indexes structure for table problem
-- ----------------------------
CREATE INDEX "idx_problem_key_lower" ON "didaoj"."problem" USING btree (
  lower(key::text) COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
CREATE INDEX "idx_problem_title_trgm" ON "didaoj"."problem" USING gin (
  "title" COLLATE "pg_catalog"."default" "public"."gin_trgm_ops"
);

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
CREATE INDEX "idx_user_username_lower" ON "didaoj"."user" USING btree (
  lower(username::text) COLLATE "pg_catalog"."default" "pg_catalog"."text_ops" ASC NULLS LAST
);
