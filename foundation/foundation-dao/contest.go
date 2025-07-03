package foundationdao

import (
	"context"
	"errors"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	"time"
)

type ContestDao struct {
	db *gorm.DB
}

var singletonContestDao = singleton.Singleton[ContestDao]{}

func GetContestDao() *ContestDao {
	return singletonContestDao.GetInstance(
		func() *ContestDao {
			dao := &ContestDao{}
			db := metamysql.GetSubsystem().GetClient("didaoj")
			dao.db = db.Model(&foundationmodel.Contest{})
			return dao
		},
	)
}

func (d *ContestDao) CheckContestEditAuth(ctx context.Context, id int, userId int) (bool, error) {
	var dummy int
	err := d.db.WithContext(ctx).
		Raw(
			`
			SELECT 1
			FROM contest c
			LEFT JOIN contest_member_auth a
			  ON a.id = c.id AND a.user_id = ?
			WHERE c.id = ? AND (c.inserter = ? OR a.user_id IS NOT NULL)
			LIMIT 1
		`, userId, id, userId,
		).
		Scan(&dummy).Error
	if err != nil {
		return false, err
	}
	return dummy == 1, nil
}

func (d *ContestDao) GetContestViewLock(ctx context.Context, id int) (*foundationview.ContestViewLock, error) {
	var contest foundationview.ContestViewLock
	err := d.db.WithContext(ctx).
		Select("id, inserter, start_time, end_time, type, always_lock, lock_rank_duration").
		Where("id = ?", id).
		Take(&contest).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &contest, nil
}

func (d *ContestDao) GetContest(ctx context.Context, id int) (*foundationview.ContestDetail, error) {
	var result foundationview.ContestDetail
	err := d.db.WithContext(ctx).
		Table("contest AS c").
		Select(
			`
			c.id, c.title, c.description, c.notification, c.start_time, c.end_time,
			c.inserter, c.modifier, c.insert_time, c.modify_time, c.password, c.private,
			u1.username AS inserter_username, u1.nickname AS inserter_nickname,
			u2.username AS modifier_username, u2.nickname AS modifier_nickname
		`,
		).
		Joins("LEFT JOIN user AS u1 ON c.inserter = u1.id").
		Joins("LEFT JOIN user AS u2 ON c.modifier = u2.id").
		Where("c.id = ?", id).
		Take(&result).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &result, nil
}

func (d *ContestDao) GetContestEdit(ctx context.Context, id int) (*foundationview.ContestDetailEdit, error) {
	var result foundationview.ContestDetailEdit
	err := d.db.WithContext(ctx).
		Table("contest AS c").
		Select(
			`
			c.id, c.title, c.description, c.notification, c.start_time, c.end_time,
			c.inserter, c.modifier, c.insert_time, c.modify_time, c.password, c.private,
			c.always_lock, c.lock_rank_duration, c.type, c.score_type, c.discuss_type,
			u1.username AS inserter_username, u1.nickname AS inserter_nickname,
			u2.username AS modifier_username, u2.nickname AS modifier_nickname
		`,
		).
		Joins("LEFT JOIN user AS u1 ON c.inserter = u1.id").
		Joins("LEFT JOIN user AS u2 ON c.modifier = u2.id").
		Where("c.id = ?", id).
		Scan(&result).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, metaerror.Wrap(err, "find contest error")
	}
	return &result, nil
}

func (d *ContestDao) GetContestViewRank(ctx context.Context, id int) (*foundationview.ContestRankDetail, error) {
	var contest foundationview.ContestRankDetail
	if err := d.db.WithContext(ctx).
		Model(&foundationmodel.Contest{}).
		Select("id, title, start_time, end_time, lock_rank_duration, always_lock").
		Where("id = ?", id).
		First(&contest).Error; err != nil {
		return nil, err
	}
	if err := d.db.WithContext(ctx).
		Table("contest_member_ignore").
		Where("id = ?", id).
		Pluck("user_id", &contest.MembersIgnore).Error; err != nil {
		return nil, err
	}
	return &contest, nil
}

func (d *ContestDao) GetProblemAttemptInfo(
	ctx context.Context,
	contestId int,
	problemIds []int,
	startTime *time.Time,
	endTime *time.Time,
) ([]*foundationview.ProblemAttemptInfo, error) {
	db := d.db.WithContext(ctx).
		Model(&foundationmodel.JudgeJob{}).
		Select(
			"problem_id AS id",
			"COUNT(*) AS attempt",
			"SUM(CASE WHEN status = ? THEN 1 ELSE 0 END) AS accept", foundationjudge.JudgeStatusAC,
		).
		Where("problem_id IN ?", problemIds).
		Where("contest_id = ?", contestId)
	if startTime != nil {
		db = db.Where("insert_time >= ?", *startTime)
	}
	if endTime != nil {
		db = db.Where("insert_time <= ?", *endTime)
	}
	var results []*foundationview.ProblemAttemptInfo
	if err := db.Group("problem_id").
		Scan(&results).Error; err != nil {
		return nil, err
	}
	return results, nil
}

func (d *ContestDao) GetContestList(
	ctx context.Context,
	title string,
	userId int,
	page int,
	pageSize int,
) ([]*foundationview.ContestList, int, error) {
	var list []*foundationview.ContestList
	var total int64

	db := d.db.WithContext(ctx).Model(&foundationmodel.Contest{})
	if title != "" {
		db = db.Where("title LIKE ?", "%"+title+"%")
	}
	if userId > 0 {
		db = db.Where("inserter = ?", userId)
	}
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to count contes")
	}
	err := db.
		Select("id", "title", "start_time", "end_time", "inserter", "private").
		Order("start_time DESC, id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&list).Error
	if err != nil {
		return nil, 0, metaerror.Wrap(err, "failed to find contes")
	}
	return list, int(total), nil
}

func (d *ContestDao) GetContestTitle(ctx context.Context, id int) (*string, error) {
	var title string
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.Contest{}).
		Where("id = ?", id).
		Pluck("title", &title).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, metaerror.Wrap(err, "find contest title error")
	}
	if title == "" {
		return nil, nil
	}
	return &title, nil
}

func (d *ContestDao) UpdateContest(
	ctx context.Context,
	contest *foundationmodel.Contest,
	contestProblems []*foundationmodel.ContestProblem,
	languages []string,
	authors []int,
	members []int,
	memberAuths []int,
	memberIgnores []int,
	memberVolunteers []int,
) error {
	if contest == nil {
		return metaerror.New("contest is nil")
	}
	err := d.db.WithContext(ctx).Transaction(
		func(tx *gorm.DB) error {
			// 更新 contest 主表
			txResult := tx.Model(contest).
				Where("id = ?", contest.Id).
				Select(
					"title", "description", "notification", "start_time", "end_time",
					"private", "password", "lock_rank_duration", "always_lock", "submit_anytime",
					"modifier", "modify_time",
				).
				Updates(contest)
			if txResult.Error != nil {
				return metaerror.Wrap(txResult.Error, "update contest")
			}
			if txResult.RowsAffected == 0 {
				return metaerror.New("no contest updated: record may not exist")
			}

			// contestProblems
			if len(contestProblems) > 0 {
				var problemIds []int
				for _, cp := range contestProblems {
					cp.Id = contest.Id
					problemIds = append(problemIds, cp.ProblemId)
				}
				if err := tx.Model(&foundationmodel.ContestProblem{}).Where(
					"id = ? AND problem_id NOT IN ?",
					contest.Id,
					problemIds,
				).
					Delete(&foundationmodel.ContestProblem{}).Error; err != nil {
					return metaerror.Wrap(err, "delete outdated contest problems")
				}
				if err := tx.Model(&foundationmodel.ContestProblem{}).Clauses(
					clause.OnConflict{
						Columns:   []clause.Column{{Name: "id"}, {Name: "problem_id"}},
						DoUpdates: clause.AssignmentColumns([]string{"index"}),
					},
				).Create(&contestProblems).Error; err != nil {
					return metaerror.Wrap(err, "upsert contest problems")
				}
			} else {
				// 如果没有 contestProblems，删除所有相关的记录
				if err := tx.Model(&foundationmodel.ContestProblem{}).Where("id = ?", contest.Id).
					Delete(&foundationmodel.ContestProblem{}).Error; err != nil {
					return metaerror.Wrap(err, "delete all contest problems")
				}
			}

			// languages
			if len(languages) > 0 {
				var langModels []*foundationmodel.ContestLanguage
				for _, lang := range languages {
					langModels = append(
						langModels, &foundationmodel.ContestLanguage{
							Id:       contest.Id,
							Language: lang,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestLanguage{}).Where(
					"id = ? AND language NOT IN ?",
					contest.Id,
					languages,
				).
					Delete(&foundationmodel.ContestLanguage{}).Error; err != nil {
					return metaerror.Wrap(err, "delete outdated languages")
				}
				if err := tx.Clauses(
					clause.OnConflict{
						Columns:   []clause.Column{{Name: "id"}, {Name: "language"}},
						DoNothing: true,
					},
				).Create(&langModels).Error; err != nil {
					return metaerror.Wrap(err, "upsert languages")
				}
			} else {
				// 如果没有 languages，删除所有相关的记录
				if err := tx.Model(&foundationmodel.ContestLanguage{}).Where("id = ?", contest.Id).
					Delete(&foundationmodel.ContestLanguage{}).Error; err != nil {
					return metaerror.Wrap(err, "delete all languages")
				}
			}

			// authors
			if len(authors) > 0 {
				var authorModels []*foundationmodel.ContestMemberAuthor
				for _, uid := range authors {
					authorModels = append(
						authorModels, &foundationmodel.ContestMemberAuthor{
							Id:     contest.Id,
							UserId: uid,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMemberAuthor{}).Where(
					"id = ? AND user_id NOT IN ?",
					contest.Id,
					authors,
				).
					Delete(&foundationmodel.ContestMemberAuthor{}).Error; err != nil {
					return metaerror.Wrap(err, "delete outdated authors")
				}
				if err := tx.Clauses(
					clause.OnConflict{
						Columns:   []clause.Column{{Name: "id"}, {Name: "user_id"}},
						DoNothing: true,
					},
				).Create(&authorModels).Error; err != nil {
					return metaerror.Wrap(err, "upsert authors")
				}
			} else {
				// 如果没有 authors，删除所有相关的记录
				if err := tx.Model(&foundationmodel.ContestMemberAuthor{}).Where("id = ?", contest.Id).
					Delete(&foundationmodel.ContestMemberAuthor{}).Error; err != nil {
					return metaerror.Wrap(err, "delete all authors")
				}
			}

			// members
			if len(members) > 0 {
				var memberModels []*foundationmodel.ContestMember
				for _, uid := range members {
					memberModels = append(
						memberModels, &foundationmodel.ContestMember{
							Id:     contest.Id,
							UserId: uid,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMember{}).Where(
					"id = ? AND user_id NOT IN ?",
					contest.Id,
					members,
				).
					Delete(&foundationmodel.ContestMember{}).Error; err != nil {
					return metaerror.Wrap(err, "delete outdated members")
				}
				if err := tx.Clauses(
					clause.OnConflict{
						Columns:   []clause.Column{{Name: "id"}, {Name: "user_id"}},
						DoNothing: true,
					},
				).Create(&memberModels).Error; err != nil {
					return metaerror.Wrap(err, "upsert members")
				}
			} else {
				// 如果没有 members，删除所有相关的记录
				if err := tx.Model(&foundationmodel.ContestMember{}).Where("id = ?", contest.Id).
					Delete(&foundationmodel.ContestMember{}).Error; err != nil {
					return metaerror.Wrap(err, "delete all members")
				}
			}

			// memberAuths
			if len(memberAuths) > 0 {
				var authModels []*foundationmodel.ContestMemberAuth
				for _, uid := range memberAuths {
					authModels = append(
						authModels, &foundationmodel.ContestMemberAuth{
							Id:     contest.Id,
							UserId: uid,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMemberAuth{}).Where(
					"id = ? AND user_id NOT IN ?",
					contest.Id,
					memberAuths,
				).
					Delete(&foundationmodel.ContestMemberAuth{}).Error; err != nil {
					return metaerror.Wrap(err, "delete outdated memberAuths")
				}
				if err := tx.Clauses(
					clause.OnConflict{
						Columns:   []clause.Column{{Name: "id"}, {Name: "user_id"}},
						DoNothing: true,
					},
				).Create(&authModels).Error; err != nil {
					return metaerror.Wrap(err, "upsert memberAuths")
				}
			} else {
				// 如果没有 memberAuths，删除所有相关的记录
				if err := tx.Model(&foundationmodel.ContestMemberAuth{}).Where("id = ?", contest.Id).
					Delete(&foundationmodel.ContestMemberAuth{}).Error; err != nil {
					return metaerror.Wrap(err, "delete all memberAuths")
				}
			}

			// memberIgnores
			if len(memberIgnores) > 0 {
				var ignoreModels []*foundationmodel.ContestMemberIgnore
				for _, uid := range memberIgnores {
					ignoreModels = append(
						ignoreModels, &foundationmodel.ContestMemberIgnore{
							Id:     contest.Id,
							UserId: uid,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMemberIgnore{}).Where(
					"id = ? AND user_id NOT IN ?",
					contest.Id,
					memberIgnores,
				).
					Delete(&foundationmodel.ContestMemberIgnore{}).Error; err != nil {
					return metaerror.Wrap(err, "delete outdated ignores")
				}
				if err := tx.Clauses(
					clause.OnConflict{
						Columns:   []clause.Column{{Name: "id"}, {Name: "user_id"}},
						DoNothing: true,
					},
				).Create(&ignoreModels).Error; err != nil {
					return metaerror.Wrap(err, "upsert ignores")
				}
			} else {
				// 如果没有 memberIgnores，删除所有相关的记录
				if err := tx.Model(&foundationmodel.ContestMemberIgnore{}).Where("id = ?", contest.Id).
					Delete(&foundationmodel.ContestMemberIgnore{}).Error; err != nil {
					return metaerror.Wrap(err, "delete all ignores")
				}
			}

			// memberVolunteers
			if len(memberVolunteers) > 0 {
				var volunteerModels []*foundationmodel.ContestMemberVolunteer
				for _, uid := range memberVolunteers {
					volunteerModels = append(
						volunteerModels, &foundationmodel.ContestMemberVolunteer{
							Id:     contest.Id,
							UserId: uid,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMemberVolunteer{}).Where(
					"id = ? AND user_id NOT IN ?",
					contest.Id,
					memberVolunteers,
				).
					Delete(&foundationmodel.ContestMemberVolunteer{}).Error; err != nil {
					return metaerror.Wrap(err, "delete outdated volunteers")
				}
				if err := tx.Clauses(
					clause.OnConflict{
						Columns:   []clause.Column{{Name: "id"}, {Name: "user_id"}},
						DoNothing: true,
					},
				).Create(&volunteerModels).Error; err != nil {
					return metaerror.Wrap(err, "upsert volunteers")
				}
			} else {
				// 如果没有 memberVolunteers，删除所有相关的记录
				if err := tx.Model(&foundationmodel.ContestMemberVolunteer{}).Where("id = ?", contest.Id).
					Delete(&foundationmodel.ContestMemberVolunteer{}).Error; err != nil {
					return metaerror.Wrap(err, "delete all volunteers")
				}
			}

			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed")
	}
	return nil
}

func (d *ContestDao) InsertContest(
	ctx context.Context,
	contest *foundationmodel.Contest,
	contestProblems []*foundationmodel.ContestProblem,
	languages []string,
	authors []int,
	members []int,
	memberAuths []int,
	memberIgnores []int,
	memberVolunteers []int,
) error {
	if contest == nil {
		return metaerror.New("contest is nil")
	}
	db := d.db.WithContext(ctx)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Model(contest).Create(contest).Error; err != nil {
				return metaerror.Wrap(err, "insert contest")
			}
			if len(contestProblems) > 0 {
				for _, c := range contestProblems {
					c.Id = contest.Id
				}
				if err := tx.Create(contestProblems).Error; err != nil {
					return metaerror.Wrap(err, "insert contest problems")
				}
			}
			if len(languages) > 0 {
				var contestLanguages []*foundationmodel.ContestLanguage
				for _, lang := range languages {
					contestLanguages = append(
						contestLanguages, &foundationmodel.ContestLanguage{
							Id:       contest.Id,
							Language: lang,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestLanguage{}).
					Create(contestLanguages).Error; err != nil {
					return metaerror.Wrap(err, "insert contest languages")
				}
			}
			if len(authors) > 0 {
				var contestAuthors []*foundationmodel.ContestMemberAuthor
				for _, authorId := range authors {
					contestAuthors = append(
						contestAuthors, &foundationmodel.ContestMemberAuthor{
							Id:     contest.Id,
							UserId: authorId,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMemberAuthor{}).
					Create(contestAuthors).Error; err != nil {
					return metaerror.Wrap(err, "insert contest authors")
				}
			}
			if len(members) > 0 {
				var contestMembers []*foundationmodel.ContestMember
				for _, memberId := range members {
					contestMembers = append(
						contestMembers, &foundationmodel.ContestMember{
							Id:     contest.Id,
							UserId: memberId,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMember{}).
					Create(contestMembers).Error; err != nil {
					return metaerror.Wrap(err, "insert contest members")
				}
			}
			if len(memberAuths) > 0 {
				var contestMemberAuths []*foundationmodel.ContestMemberAuthor
				for _, memberAuthId := range memberAuths {
					contestMemberAuths = append(
						contestMemberAuths, &foundationmodel.ContestMemberAuthor{
							Id:     contest.Id,
							UserId: memberAuthId,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMemberAuthor{}).
					Create(contestMemberAuths).Error; err != nil {
					return metaerror.Wrap(err, "insert contest member authors")
				}
			}
			if len(memberIgnores) > 0 {
				var contestMemberIgnores []*foundationmodel.ContestMemberIgnore
				for _, memberIgnoreId := range memberIgnores {
					contestMemberIgnores = append(
						contestMemberIgnores, &foundationmodel.ContestMemberIgnore{
							Id:     contest.Id,
							UserId: memberIgnoreId,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMemberIgnore{}).
					Create(contestMemberIgnores).Error; err != nil {
					return metaerror.Wrap(err, "insert contest member ignores")
				}
			}
			if len(memberVolunteers) > 0 {
				var contestMemberVolunteers []*foundationmodel.ContestMemberVolunteer
				for _, memberVolunteerId := range memberVolunteers {
					contestMemberVolunteers = append(
						contestMemberVolunteers, &foundationmodel.ContestMemberVolunteer{
							Id:     contest.Id,
							UserId: memberVolunteerId,
						},
					)
				}
				if err := tx.Model(&foundationmodel.ContestMemberVolunteer{}).
					Create(contestMemberVolunteers).Error; err != nil {
					return metaerror.Wrap(err, "insert contest member volunteers")
				}
			}

			return nil
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "transaction failed")
	}
	return nil
}
