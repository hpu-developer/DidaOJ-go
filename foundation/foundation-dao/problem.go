package foundationdao

import (
	"context"
	"errors"
	"fmt"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	metatime "meta/meta-time"
	"meta/singleton"
	"strconv"
	"strings"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ProblemDao struct {
	db *gorm.DB
}

var singletonProblemDao = singleton.Singleton[ProblemDao]{}

func GetProblemDao() *ProblemDao {
	return singletonProblemDao.GetInstance(
		func() *ProblemDao {
			dao := &ProblemDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *ProblemDao) GetProblemList(
	ctx context.Context,
	oj string, title string, tags []int, private bool,
	userId int, hasAuth bool,
	page int,
	pageSize int,
) ([]*foundationview.ProblemViewList, int, error) {

	db := d.db.WithContext(ctx).Model(&foundationmodel.Problem{})
	db = db.Select("problem.id AS id, problem.key AS key, problem.title, problem.accept, problem.attempt")

	if !hasAuth {
		if userId > 0 {
			db = db.Where(
				`
				problem.private = FALSE
				OR problem.inserter = ?
				OR EXISTS (
					SELECT 1 FROM problem_member pm
					WHERE pm.id = problem.id AND pm.user_id = ?
				)
				OR EXISTS (
					SELECT 1 FROM problem_member_auth pma
					WHERE pma.id = problem.id AND pma.user_id = ?
				)
			`, userId, userId, userId,
			)
		} else {
			db = db.Where("problem.private = FALSE")
		}
	} else if private {
		db = db.Where("problem.private = TRUE")
	}

	db = db.Joins("LEFT JOIN problem_remote r ON r.problem_id = problem.id")
	if oj == "didaoj" {
		db = db.Where("r.origin_oj IS NULL")
	} else if oj != "" {
		db = db.Where("r.origin_oj = ?", oj)
	}

	if title != "" {
		// ILIKE 支持不区分大小写的匹配（Postgres）
		db = db.Where("problem.title ILIKE ?", "%"+title+"%")
	}

	if len(tags) > 0 {
		db = db.Joins("JOIN problem_tag pt ON pt.id = problem.id").
			Where("pt.tag_id IN ?", tags).
			Group("problem.id, problem.key, problem.title, problem.accept, problem.attempt")
	}

	if page < 1 {
		page = 1
	}
	offset := (page - 1) * pageSize

	var total int64
	countDB := db.Session(&gorm.Session{})
	if len(tags) > 0 {
		if err := countDB.Select("COUNT(DISTINCT problem.id)").Count(&total).Error; err != nil {
			return nil, 0, metaerror.Wrap(err, "count failed")
		}
	} else {
		if err := countDB.Count(&total).Error; err != nil {
			return nil, 0, metaerror.Wrap(err, "count failed")
		}
	}

	var list []*foundationview.ProblemViewList
	if err := db.Order("problem.id ASC").
		Offset(offset).
		Limit(pageSize).
		Find(&list).Error; err != nil {
		return nil, 0, metaerror.Wrap(err, "find failed")
	}
	return list, int(total), nil
}

func (d *ProblemDao) GetProblemView(
	ctx context.Context, id int, userId int, hasAuth bool,
) (*foundationview.Problem, error) {
	db := d.db.WithContext(ctx).Table("problem AS p").
		Select(
			`
			p.*,
			u1.username AS inserter_username, u1.nickname AS inserter_nickname,
			u2.username AS modifier_username, u2.nickname AS modifier_nickname,
			r.origin_oj, r.origin_id, r.origin_url, r.origin_author
		`,
		).
		Joins(`LEFT JOIN "user" u1 ON u1.id = p.inserter`).
		Joins(`LEFT JOIN "user" u2 ON u2.id = p.modifier`).
		Joins(`LEFT JOIN problem_remote r ON r.problem_id = p.id`).
		Where("p.id = ?", id)

	if !hasAuth {
		if userId > 0 {
			db = db.Where(
				`
				p.private = FALSE OR
				p.inserter = ? OR
				EXISTS (
					SELECT 1 FROM problem_member pm WHERE pm.id = p.id AND pm.user_id = ?
				) OR
				EXISTS (
					SELECT 1 FROM problem_member_auth pam WHERE pam.id = p.id AND pam.user_id = ?
				)
			`, userId, userId, userId,
			)
		} else {
			db = db.Where("p.private = FALSE")
		}
	}

	var problem foundationview.Problem
	if err := db.First(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem with remote info error")
	}
	return &problem, nil
}

func (d *ProblemDao) CheckProblemEditAuth(ctx context.Context, problemId int, userId int) (bool, error) {
	var exists int
	err := d.db.WithContext(ctx).Raw(
		`SELECT 1
		FROM problem p
		LEFT JOIN problem_member_auth pa ON p.id = pa.id AND pa.user_id = ?
		WHERE p.id = ? AND (p.inserter = ? OR pa.user_id IS NOT NULL)
		LIMIT 1
	`, userId, problemId, userId,
	).Scan(&exists).Error

	if err != nil {
		return false, metaerror.Wrap(err, "check edit permission failed")
	}
	return exists == 1, nil
}

func (d *ProblemDao) CheckProblemSubmitAuth(ctx context.Context, problemId int, userId int) (bool, error) {
	var exists int
	err := d.db.WithContext(ctx).Raw(
		`
		SELECT 1
		FROM problem p
		LEFT JOIN problem_member m ON p.id = m.id AND m.user_id = ?
		LEFT JOIN problem_member_auth a ON p.id = a.id AND a.user_id = ?
		WHERE p.id = ?
		  AND (
		    p.private = FALSE
		    OR p.inserter = ?
		    OR m.user_id IS NOT NULL
		    OR a.user_id IS NOT NULL
		  )
		LIMIT 1;
	`, userId, userId, problemId, userId,
	).Scan(&exists).Error
	if err != nil {
		return false, metaerror.Wrap(err, "check submit permission failed")
	}
	return exists == 1, nil
}

func (d *ProblemDao) GetProblemViewAuth(ctx context.Context, id string) (*foundationview.ProblemViewAuth, error) {
	var problem foundationview.ProblemViewAuth
	tx := d.db.WithContext(ctx)
	if err := tx.
		Select("id", "inserter", "private").
		Where("id = ?", id).
		Take(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem error")
	}
	if err := tx.
		Model(&foundationmodel.ProblemMember{}).
		Where("id = ?", id).
		Pluck("user_id", &problem.Members).Error; err != nil {
		return nil, metaerror.Wrap(err, "find problem members error")
	}
	if err := tx.
		Model(&foundationmodel.ProblemMemberAuth{}).
		Where("id = ?", id).
		Pluck("user_id", &problem.AuthMembers).Error; err != nil {
		return nil, metaerror.Wrap(err, "find problem auth members error")
	}
	return &problem, nil
}

func (d *ProblemDao) HasProblem(ctx context.Context, id int) (bool, error) {
	var dummy int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("1").
		Where("id = ?", id).
		Limit(1).
		Scan(&dummy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return dummy == 1, nil
}

func (d *ProblemDao) HasProblemByKey(ctx context.Context, key string) (bool, error) {
	var dummy int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("1").
		Where("key = ?", key).
		Limit(1).
		Scan(&dummy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}
	return dummy == 1, nil
}

func (d *ProblemDao) HasProblemTitle(ctx context.Context, title string) (bool, error) {
	var dummy int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("1").
		Where("title = ?", title).
		Limit(1).
		Scan(&dummy).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, metaerror.Wrap(err, "check problem title error")
	}
	return dummy == 1, nil
}

func (d *ProblemDao) GetProblemIdByKey(ctx context.Context, key string) (int, error) {
	var result struct {
		Id int `gorm:"column:id"`
	}
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("id").
		Where("LOWER(key) = LOWER(?)", key).
		Take(&result).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, metaerror.New("problem %s not found", key)
		}
		return 0, metaerror.Wrap(err, "find problem id by key %s failed", key)
	}
	return result.Id, nil
}

func (d *ProblemDao) GetProblemIdsByKey(ctx context.Context, problemKeys []string) ([]int, error) {
	if len(problemKeys) == 0 {
		return nil, nil
	}
	lowerKeys := make([]string, len(problemKeys))
	for i, k := range problemKeys {
		lowerKeys[i] = strings.ToLower(k)
	}
	var ids []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("id").
		Where("LOWER(key) IN ?", lowerKeys).
		Pluck("id", &ids).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "find problem ids by keys failed")
	}
	return ids, nil
}

func (d *ProblemDao) CheckProblemIdViewByKey(
	ctx context.Context, problemKey string,
	userId int, hasAuth bool,
) (
	int,
	error,
) {
	db := d.db.WithContext(ctx).Table("problem AS p").
		Select(`p.id`).
		Where("p.key = ?", problemKey)

	if !hasAuth {
		if userId > 0 {
			db = db.Where(
				`
				p.private = FALSE OR
				p.inserter = ? OR
				EXISTS (
					SELECT 1 FROM problem_member pm WHERE pm.id = p.id AND pm.user_id = ?
				) OR
				EXISTS (
					SELECT 1 FROM problem_member_auth pam WHERE pam.id = p.id AND pam.user_id = ?
				)
			`, userId, userId, userId,
			)
		} else {
			db = db.Where("p.private = FALSE")
		}
	}
	var problemId int
	err := db.Take(&problemId).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	return problemId, err
}

func (d *ProblemDao) GetProblemTitle(ctx context.Context, id int) (*string, error) {
	var problem foundationmodel.Problem
	err := d.db.WithContext(ctx).
		Select("title").
		Where("id = ?", id).
		Take(&problem).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &problem.Title, err
}

func (d *ProblemDao) GetProblemTitles(
	ctx context.Context,
	userId int,
	hasAuth bool,
	ids []int,
) ([]*foundationview.ProblemViewTitle, error) {
	query := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("id", "key", "title").
		Where("id IN ?", ids)
	if !hasAuth {
		if userId > 0 {
			query = query.Where(
				`
				private = FALSE OR
				inserter = ? OR
				EXISTS (
					SELECT 1 FROM problem_member pm
					WHERE pm.problem_id = problems.id AND pm.user_id = ?
				) OR
				EXISTS (
					SELECT 1 FROM problem_member_auth pam
					WHERE pam.problem_id = problems.id AND pam.user_id = ?
				)
			`, userId, userId, userId,
			)
		} else {
			query = query.Where("private = FALSE")
		}
	}
	var titles []*foundationview.ProblemViewTitle
	err := query.Find(&titles).Error
	return titles, err
}

func (d *ProblemDao) GetProblemViewForLocalJudge(ctx context.Context, id int) (
	*foundationview.ProblemForLocalJudge,
	error,
) {
	db := d.db.WithContext(ctx).Table("problem AS p").
		Select(
			`
			p.id, p.judge_type, p.time_limit, p.memory_limit,
			r.judge_md5
		`,
		).
		Joins(`LEFT JOIN problem_local r ON r.problem_id = p.id`).
		Where("p.id = ?", id)
	var problem foundationview.ProblemForLocalJudge
	if err := db.First(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem with remote info error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemViewForRemoteJudge(ctx context.Context, id int) (
	*foundationview.ProblemForRemoteJudge,
	error,
) {
	db := d.db.WithContext(ctx).Table("problem AS p").
		Select(
			`
			p.id,
			r.origin_oj, r.origin_id
		`,
		).
		Joins(`LEFT JOIN problem_remote r ON r.problem_id = p.id`).
		Where("p.id = ?", id)
	var problem foundationview.ProblemForRemoteJudge
	if err := db.First(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem with remote info error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemViewJudgeData(ctx context.Context, id int) (*foundationview.ProblemJudgeData, error) {
	db := d.db.WithContext(ctx).Table("problem AS p").
		Select(
			`
			p.id, p.key, p.title, p.judge_type,p.inserter, p.insert_time, p.modifier, p.modify_time,
			u1.username AS inserter_username, u1.nickname AS inserter_nickname,
			u2.username AS modifier_username, u2.nickname AS modifier_nickname,
			r.judge_md5, r.judge_job
		`,
		).
		Joins(`LEFT JOIN "user" u1 ON u1.id = p.inserter`).
		Joins(`LEFT JOIN "user" u2 ON u2.id = p.modifier`).
		Joins(`LEFT JOIN problem_local r ON r.problem_id = p.id`).
		Where("p.id = ?", id)
	var problem foundationview.ProblemJudgeData
	if err := db.First(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem with remote info error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemViewJudgeDataByKey(ctx context.Context, key string) (
	*foundationview.ProblemJudgeData,
	error,
) {
	db := d.db.WithContext(ctx).Table("problem AS p").
		Select(
			`
			p.id, p.key, p.title, p.judge_type,p.inserter, p.insert_time, p.modifier, p.modify_time,
			u1.username AS inserter_username, u1.nickname AS inserter_nickname,
			u2.username AS modifier_username, u2.nickname AS modifier_nickname,
			r.judge_md5, r.judge_job
		`,
		).
		Joins(`LEFT JOIN "user" u1 ON u1.id = p.inserter`).
		Joins(`LEFT JOIN "user" u2 ON u2.id = p.modifier`).
		Joins(`LEFT JOIN problem_local r ON r.problem_id = p.id`).
		Where("p.key = ?", key)
	var problem foundationview.ProblemJudgeData
	if err := db.First(&problem).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem with remote info error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemViewApproveJudge(ctx context.Context, id int) (
	*foundationview.ProblemViewApproveJudge,
	error,
) {
	db := d.db.WithContext(ctx).
		Table("problem as p").
		Select("p.id", "pr.origin_oj", "pr.origin_id").
		Joins("LEFT JOIN problem_remote as pr ON p.id = pr.problem_id").
		Where("p.id = ?", id)
	var problem foundationview.ProblemViewApproveJudge
	err := db.First(&problem).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find problem approve judge error")
	}
	return &problem, nil
}

func (d *ProblemDao) GetProblemJudgeMd5(ctx context.Context, id string) (*string, error) {
	var result struct {
		JudgeMd5 *string `gorm:"column:judge_md5"`
	}
	err := d.db.WithContext(ctx).Model(&foundationmodel.ProblemLocal{}).
		Select("judge_md5").
		Where("id = ?", id).
		Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return result.JudgeMd5, err
}

func (d *ProblemDao) GetProblemDescription(ctx context.Context, id int) (*string, error) {
	var result struct {
		Description string `gorm:"column:description"`
	}
	err := d.db.WithContext(ctx).Model(&foundationmodel.Problem{}).
		Select("description").
		Where("id = ?", id).
		Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &result.Description, err
}

func (d *ProblemDao) GetProblemListTitle(ctx context.Context, ids []string) (
	[]*foundationview.ProblemViewTitle,
	error,
) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*foundationview.ProblemViewTitle
	err := d.db.WithContext(ctx).
		Select("id", "title").
		Where("id IN ?", ids).
		Find(&list).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get problem list title error")
	}
	return list, nil
}

func (d *ProblemDao) FilterValidProblemIds(ctx context.Context, ids []int) ([]int, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var validIds []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Select("id").
		Where("id IN ?", ids).
		Pluck("id", &validIds).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "find problem error")
	}
	return validIds, nil
}

func (d *ProblemDao) SelectProblemViewList(
	ctx context.Context,
	ids []int,
	needAttempt bool,
) ([]*foundationview.ProblemViewList, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var list []*foundationview.ProblemViewList
	fields := []string{"id", "key", "title"}
	if needAttempt {
		fields = append(fields, "accept", "attempt")
	}
	err := d.db.WithContext(ctx).
		Select(fields).
		Where("id IN ?", ids).
		Find(&list).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "get problem error")
	}
	return list, nil
}

func (d *ProblemDao) UpdateProblem(
	ctx context.Context,
	problemId int,
	problem *foundationmodel.Problem,
	tags []string,
) error {
	err := d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			var tagIds []int
			for _, tagName := range tags {
				id, err := GetTagDao().InsertTagWithDb(tx, tagName)
				if err != nil {
					return err
				}
				tagIds = append(tagIds, id)
			}
			updateData := map[string]interface{}{
				"title":        problem.Title,
				"description":  problem.Description,
				"time_limit":   problem.TimeLimit,
				"memory_limit": problem.MemoryLimit,
				"source":       problem.Source,
				"modifier":     problem.Modifier,
				"modify_time":  problem.ModifyTime,
				"private":      problem.Private,
			}
			txRes := tx.Model(&foundationmodel.Problem{}).
				Where("id = ?", problemId).
				Updates(updateData)
			if txRes.Error != nil {
				return txRes.Error
			}
			if txRes.RowsAffected == 0 {
				return metaerror.New("problem not found")
			}
			if err := GetProblemTagDao().UpdateProblemTagsByDb(tx, problemId, tagIds); err != nil {
				return err
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed to update problem")
	}
	return nil
}

func (d *ProblemDao) UpdateProblemDescription(
	ctx context.Context,
	id int,
	description string,
) error {
	nowTime := metatime.GetTimeNow()
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Problem{}).
		Where("id = ?", id).
		Updates(
			map[string]interface{}{
				"description": description,
				"modify_time": nowTime,
			},
		).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return metaerror.New("problem not found")
		}
		return metaerror.Wrap(err, "failed to update problem description")
	}
	return nil
}

func (d *ProblemDao) UpdateProblemJudgeInfo(
	ctx context.Context,
	id int,
	judgeType foundationjudge.JudgeType,
	md5 string,
	jobConfig foundationjudge.JudgeJobConfig,
) error {
	nowTime := metatime.GetTimeNow()
	err := d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			err := tx.Model(&foundationmodel.Problem{}).
				Where("id = ?", id).
				Updates(
					map[string]interface{}{
						"judge_type":  judgeType,
						"modify_time": nowTime,
					},
				).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return metaerror.New("problem not found")
				}
				return metaerror.Wrap(err, "failed to update problem judge info")
			}
			err = tx.Model(&foundationmodel.ProblemLocal{}).
				Where("problem_id = ?", id).
				Updates(
					map[string]interface{}{
						"judge_md5": md5,
						"judge_job": jobConfig,
					},
				).Error
			if err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return metaerror.New("problem local not found")
				}
				return metaerror.Wrap(err, "failed to update problem local judge md5")
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed to update problem judge info")
	}
	return nil
}

func (d *ProblemDao) UpdateProblemCrawl(
	ctx context.Context,
	problemKey string,
	problem *foundationmodel.Problem,
	problemRemote *foundationmodel.ProblemRemote,
) error {
	return d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			var existing foundationmodel.Problem
			// 查找是否已存在该 problemKey，尝试加锁避免并发写入
			err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("key = ?", problemKey).
				First(&existing).Error
			var problemId int
			if err == nil {
				// 已存在，执行更新
				problem.Id = existing.Id // 保留 ID
				updates := map[string]interface{}{
					"title":        problem.Title,
					"description":  problem.Description,
					"source":       problem.Source,
					"time_limit":   problem.TimeLimit,
					"memory_limit": problem.MemoryLimit,
					"judge_type":   problem.JudgeType,
					"modifier":     problem.Modifier,
					"modify_time":  problem.ModifyTime,
				}
				if err := tx.Model(&foundationmodel.Problem{}).
					Where("id = ?", existing.Id).
					Updates(updates).Error; err != nil {
					return err
				}
				problemId = existing.Id
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 不存在，插入
				problem.Key = problemKey // 确保 key 被写入
				if err := tx.Create(problem).Error; err != nil {
					return err
				}
				problemId = problem.Id
			} else {
				// 其他错误
				return err
			}

			// 处理 problem_remote
			var remoteExisting foundationmodel.ProblemRemote
			err = tx.Clauses(clause.Locking{Strength: "UPDATE"}).
				Where("problem_id = ?", problemId).
				First(&remoteExisting).Error
			problemRemote.ProblemId = problemId // 设置关联 ID
			if err == nil {
				// 已存在，更新
				if err := tx.Model(&foundationmodel.ProblemRemote{}).
					Where("problem_id = ?", problemId).
					Updates(problemRemote).Error; err != nil {
					return err
				}
			} else if errors.Is(err, gorm.ErrRecordNotFound) {
				// 不存在，插入
				if err := tx.Create(problemRemote).Error; err != nil {
					return err
				}
			} else {
				return err
			}
			return nil
		},
	)
}

func (d *ProblemDao) InsertProblemLocal(
	ctx context.Context,
	problem *foundationmodel.Problem,
	problemLocal *foundationmodel.ProblemLocal,
	tags []string,
) error {
	if problem == nil {
		return metaerror.New("problem is nil")
	}
	db := d.db.WithContext(ctx)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			var tagIds []int
			for _, tagName := range tags {
				id, err := GetTagDao().InsertTagWithDb(tx, tagName)
				if err != nil {
					return err
				}
				tagIds = append(tagIds, id)
			}
			if err := tx.Create(problem).Error; err != nil {
				return metaerror.Wrap(err, "insert problem")
			}
			if problemLocal == nil {
				problemLocal = &foundationmodel.ProblemLocal{}
			}
			problemLocal.ProblemId = problem.Id
			if err := tx.Create(problemLocal).Error; err != nil {
				return metaerror.Wrap(err, "insert problem local")
			}
			// key 使用 problemLocal.Id 转为字符串
			problem.Key = strconv.Itoa(problemLocal.Id)
			if err := tx.Save(problem).Error; err != nil {
				return metaerror.Wrap(err, "update problem key")
			}
			if err := GetProblemTagDao().UpdateProblemTagsByDb(tx, problem.Id, tagIds); err != nil {
				return metaerror.Wrap(err, "update problem tags")
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *ProblemDao) InsertProblemRemote(
	ctx context.Context,
	problem *foundationmodel.Problem,
	problemRemote *foundationmodel.ProblemRemote,
) error {
	if problem == nil {
		return metaerror.New("problem is nil")
	}
	if problemRemote == nil {
		return metaerror.New("problemRemote is nil")
	}
	if problemRemote.OriginOj == "" || problemRemote.OriginId == "" {
		return metaerror.New("problemRemote originOj or originId is nil")
	}
	db := d.db.WithContext(ctx)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Create(problem).Error; err != nil {
				return metaerror.Wrap(err, "insert problem")
			}
			if problemRemote == nil {
				problemRemote = &foundationmodel.ProblemRemote{}
			}
			problemRemote.ProblemId = problem.Id
			if err := tx.Create(problemRemote).Error; err != nil {
				return metaerror.Wrap(err, "insert problem remote")
			}
			problem.Key = fmt.Sprintf("%s-%s", problemRemote.OriginOj, problemRemote.OriginId)
			if err := tx.Save(problem).Error; err != nil {
				return metaerror.Wrap(err, "update problem key")
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed")
	}
	return nil
}
