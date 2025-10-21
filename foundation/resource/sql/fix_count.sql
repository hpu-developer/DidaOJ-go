UPDATE problem p
JOIN (
    SELECT
        problem_id,
        COUNT(*) AS attempt,
        SUM(CASE WHEN status = 6 THEN 1 ELSE 0 END) AS accept
    FROM judge_job
    GROUP BY problem_id
) AS j ON p.id = j.problem_id
SET p.attempt = j.attempt,
    p.accept = j.accept;

UPDATE user u
JOIN (
    SELECT
        inserter AS user_id,
        COUNT(*) AS attempt,
        SUM(CASE WHEN status = 6 THEN 1 ELSE 0 END) AS accept
    FROM judge_job
    GROUP BY inserter
) AS j ON u.id = j.user_id
SET
    u.attempt = j.attempt,
    u.accept = j.accept;
