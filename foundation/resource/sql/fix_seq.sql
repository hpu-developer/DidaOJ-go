-- ----------------------------
-- 修复 didaoj.user_role 表序列
-- ----------------------------
-- 如果表中已经有数据，需要同步序列
DO
$$
DECLARE
    max_id BIGINT;
BEGIN
    -- user_role 表
    SELECT COALESCE(MAX(id), 0) INTO max_id FROM didaoj.user_role;
    -- 创建序列（如果不存在）
    CREATE SEQUENCE IF NOT EXISTS didaoj.user_role_id_seq;
    -- 关联序列到列
    ALTER TABLE didaoj.user_role ALTER COLUMN id SET DEFAULT nextval('didaoj.user_role_id_seq');
    -- 同步序列值
    PERFORM setval('didaoj.user_role_id_seq', max_id, true);
    
    -- collection 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj.collection;
    CREATE SEQUENCE IF NOT EXISTS didaoj.collection_id_seq;
    ALTER TABLE didaoj.collection ALTER COLUMN id SET DEFAULT nextval('didaoj.collection_id_seq');
    PERFORM setval('didaoj.collection_id_seq', max_id, true);

    -- contest 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj.contest;
    CREATE SEQUENCE IF NOT EXISTS didaoj.contest_id_seq;
    ALTER TABLE didaoj.contest ALTER COLUMN id SET DEFAULT nextval('didaoj.contest_id_seq');
    PERFORM setval('didaoj.contest_id_seq', max_id, true);

    -- discuss_comment 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj.discuss_comment;
    CREATE SEQUENCE IF NOT EXISTS didaoj.discuss_comment_id_seq;
    ALTER TABLE didaoj.discuss_comment ALTER COLUMN id SET DEFAULT nextval('didaoj.discuss_comment_id_seq');
    PERFORM setval('didaoj.discuss_comment_id_seq', max_id, true);

    -- discuss 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj.discuss;
    CREATE SEQUENCE IF NOT EXISTS didaoj.discuss_id_seq;
    ALTER TABLE didaoj.discuss ALTER COLUMN id SET DEFAULT nextval('didaoj.discuss_id_seq');
    PERFORM setval('didaoj.discuss_id_seq', max_id, true);

    -- judge_job 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj.judge_job;
    CREATE SEQUENCE IF NOT EXISTS didaoj.judge_job_id_seq;
    ALTER TABLE didaoj.judge_job ALTER COLUMN id SET DEFAULT nextval('didaoj.judge_job_id_seq');
    PERFORM setval('didaoj.judge_job_id_seq', max_id, true);

    -- problem 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj.problem;
    CREATE SEQUENCE IF NOT EXISTS didaoj.problem_id_seq;
    ALTER TABLE didaoj.problem ALTER COLUMN id SET DEFAULT nextval('didaoj.problem_id_seq');
    PERFORM setval('didaoj.problem_id_seq', max_id, true);

    -- problem_local 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj.problem_local;
    CREATE SEQUENCE IF NOT EXISTS didaoj.problem_local_id_seq;
    ALTER TABLE didaoj.problem_local ALTER COLUMN id SET DEFAULT nextval('didaoj.problem_local_id_seq');
    PERFORM setval('didaoj.problem_local_id_seq', max_id, true);

    -- problem_remote 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj.problem_remote;
    CREATE SEQUENCE IF NOT EXISTS didaoj.problem_remote_id_seq;
    ALTER TABLE didaoj.problem_remote ALTER COLUMN id SET DEFAULT nextval('didaoj.problem_remote_id_seq');
    PERFORM setval('didaoj.problem_remote_id_seq', max_id, true);

    -- tag 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj.tag;
    CREATE SEQUENCE IF NOT EXISTS didaoj.tag_id_seq;
    ALTER TABLE didaoj.tag ALTER COLUMN id SET DEFAULT nextval('didaoj.tag_id_seq');
    PERFORM setval('didaoj.tag_id_seq', max_id, true);

    -- user 表
    SELECT COALESCE(MAX(id),0) INTO max_id FROM didaoj."user";
    CREATE SEQUENCE IF NOT EXISTS didaoj.user_id_seq;
    ALTER TABLE didaoj."user" ALTER COLUMN id SET DEFAULT nextval('didaoj.user_id_seq');
    PERFORM setval('didaoj.user_id_seq', max_id, true);

END
$$;
