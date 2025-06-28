package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	"gorm.io/gorm"
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
	db := d.db.WithContext(ctx)
	err := db.Transaction(
		func(tx *gorm.DB) error {
			if err := tx.Model(contest).Save(contest).Error; err != nil {
				return metaerror.Wrap(err, "insert contest")
			}
			if err := tx.Model(&foundationmodel.ContestProblem{}).
				Where("id = ?", contest.Id).
				Delete(&foundationmodel.ContestProblem{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old contest problems")
			}
			if err := tx.Model(&foundationmodel.ContestLanguage{}).
				Where("id = ?", contest.Id).
				Delete(&foundationmodel.ContestLanguage{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old contest languages")
			}
			if err := tx.Model(&foundationmodel.ContestMemberAuthor{}).
				Where("id = ?", contest.Id).
				Delete(&foundationmodel.ContestMemberAuthor{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old contest authors")
			}
			if err := tx.Model(&foundationmodel.ContestMember{}).
				Where("id = ?", contest.Id).
				Delete(&foundationmodel.ContestMember{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old contest members")
			}
			if err := tx.Model(&foundationmodel.ContestMemberIgnore{}).
				Where("id = ?", contest.Id).
				Delete(&foundationmodel.ContestMemberIgnore{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old contest ignores")
			}
			if err := tx.Model(&foundationmodel.ContestMemberVolunteer{}).
				Where("id = ?", contest.Id).
				Delete(&foundationmodel.ContestMemberVolunteer{}).Error; err != nil {
				return metaerror.Wrap(err, "delete old contest volunteers")
			}
			if len(contestProblems) > 0 {
				if err := tx.Model(&foundationmodel.ContestProblem{}).
					Create(contestProblems).Error; err != nil {
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
				if err := tx.Model(&foundationmodel.ContestProblem{}).
					Create(contestProblems).Error; err != nil {
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
