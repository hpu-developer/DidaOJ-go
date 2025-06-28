package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type ContestDao struct {
	db *gorm.DB
}

var singletonContestDao = singleton.Singleton[ContestDao]{}

func GetContestDao() *ContestDao {
	return singletonContestDao.GetInstance(
		func() *ContestDao {
			dao := &ContestDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
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
			if err := tx.Model(contest).Save(contest).Error; err != nil {
				return metaerror.Wrap(err, "insert contest")
			}

			// contestProblems
			if len(contestProblems) > 0 {
				var problemIds []int
				for _, cp := range contestProblems {
					cp.Id = contest.Id
					problemIds = append(problemIds, cp.ProblemId)
				}
				if err := tx.Where("id = ? AND problem_id NOT IN ?", contest.Id, problemIds).
					Delete(&foundationmodel.ContestProblem{}).Error; err != nil {
					return metaerror.Wrap(err, "delete outdated contest problems")
				}
				if err := tx.Clauses(
					clause.OnConflict{
						Columns:   []clause.Column{{Name: "id"}, {Name: "problem_id"}},
						DoUpdates: clause.AssignmentColumns([]string{"index"}),
					},
				).Create(&contestProblems).Error; err != nil {
					return metaerror.Wrap(err, "upsert contest problems")
				}
			} else {
				// 如果没有 contestProblems，删除所有相关的记录
				if err := tx.Where("id = ?", contest.Id).
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
				if err := tx.Where("id = ? AND language NOT IN ?", contest.Id, languages).
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
				if err := tx.Where("id = ?", contest.Id).
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
				if err := tx.Where("id = ? AND user_id NOT IN ?", contest.Id, authors).
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
				if err := tx.Where("id = ?", contest.Id).
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
				if err := tx.Where("id = ? AND user_id NOT IN ?", contest.Id, members).
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
				if err := tx.Where("id = ?", contest.Id).
					Delete(&foundationmodel.ContestMember{}).Error; err != nil {
					return metaerror.Wrap(err, "delete all members")
				}
			}

			// memberAuths
			if len(memberAuths) > 0 {
				var authModels []*foundationmodel.ContestMemberAuthor
				for _, uid := range memberAuths {
					authModels = append(
						authModels, &foundationmodel.ContestMemberAuthor{
							Id:     contest.Id,
							UserId: uid,
						},
					)
				}
				if err := tx.Where("id = ? AND user_id NOT IN ?", contest.Id, memberAuths).
					Delete(&foundationmodel.ContestMemberAuthor{}).Error; err != nil {
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
				if err := tx.Where("id = ?", contest.Id).
					Delete(&foundationmodel.ContestMemberAuthor{}).Error; err != nil {
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
				if err := tx.Where("id = ? AND user_id NOT IN ?", contest.Id, memberIgnores).
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
				if err := tx.Where("id = ?", contest.Id).
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
				if err := tx.Where("id = ? AND user_id NOT IN ?", contest.Id, memberVolunteers).
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
				if err := tx.Where("id = ?", contest.Id).
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
