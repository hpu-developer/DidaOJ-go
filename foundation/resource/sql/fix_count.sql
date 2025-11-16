UPDATE problem p
SET
    attempt = j.attempt,
    accept  = j.accept
FROM (
    SELECT
        problem_id,
        COUNT(*) AS attempt,
        SUM(CASE WHEN status = 6 THEN 1 ELSE 0 END) AS accept
    FROM judge_job
    GROUP BY problem_id
) AS j
WHERE p.id = j.problem_id;

UPDATE "user" u
SET
    attempt = j.attempt,
    accept  = j.accept
FROM (
    SELECT
        inserter AS user_id,
        COUNT(*) AS attempt,
        SUM(CASE WHEN status = 6 THEN 1 ELSE 0 END) AS accept
    FROM judge_job
    GROUP BY inserter
) AS j
WHERE u.id = j.user_id;
