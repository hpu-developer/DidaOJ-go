package foundationdao

import (
	"context"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type ContestProblemDao struct {
	db *gorm.DB
}

var singletonContestProblemDao = singleton.Singleton[ContestProblemDao]{}

func GetContestProblemDao() *ContestProblemDao {
	return singletonContestProblemDao.GetInstance(
		func() *ContestProblemDao {
			dao := &ContestProblemDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *ContestProblemDao) GetProblemId(ctx context.Context, id int, index int) (int, error) {
	var problemId int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.ContestProblem{}).
		Where("id = ? AND `index` = ?", id, index).
		Pluck("problem_id", &problemId).Error // index 是保留字，建议加反引号
	if err != nil {
		return 0, err
	}
	if problemId == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return problemId, nil
}

func (d *ContestProblemDao) GetProblemIndex(ctx context.Context, id int, problemId int) (int, error) {
	var index int
	err := d.db.WithContext(ctx).Model(&foundationmodel.ContestProblem{}).
		Where("contest_id = ? AND problem_id = ?", id, problemId).
		Pluck("`index`", &index).Error // index 是保留字，建议加反引号
	if err != nil {
		return 0, err
	}
	if index == 0 {
		return 0, gorm.ErrRecordNotFound
	}
	return index, nil
}

func (d *ContestProblemDao) GetProblems(ctx context.Context, contestId int) (
	[]*foundationmodel.ContestProblem,
	error,
) {
	var results []*foundationmodel.ContestProblem
	err := d.db.WithContext(ctx).Model(&foundationmodel.ContestProblem{}).
		Where("id = ?", contestId).
		Find(&results).Error
	if err != nil {
		return nil, metaerror.Wrap(err, "find contest problems error")
	}
	return results, nil
}

func (d *ContestProblemDao) GetProblemIds(ctx context.Context, contestId int) (
	[]int,
	error,
) {
	var results []int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.ContestProblem{}).
		Where("id = ?", contestId).
		Pluck("problem_id", &results).Error // 使用 Pluck 获取指定字段的值
	if err != nil {
		return nil, err
	}
	return results, nil
}

func (d *ContestProblemDao) GetProblemsRank(ctx context.Context, id int) ([]*foundationview.ContestProblemRank, error) {
	var results []*foundationview.ContestProblemRank
	err := d.db.WithContext(ctx).Model(&foundationmodel.ContestProblem{}).
		Select("problem_id,`index`,score").
		Where("id = ?", id).
		Scan(&results).Error

	if err != nil {
		return nil, metaerror.Wrap(err)
	}
	return results, nil
}

func (d *ContestProblemDao) GetProblemsDetail(ctx context.Context, contestId int) (
	[]*foundationview.ContestProblemDetail,
	error,
) {
	var results []*foundationview.ContestProblemDetail
	err := d.db.WithContext(ctx).
		Table("contest_problem AS cp").
		Select(
			`
			cp.id,
			cp.problem_id,
			cp.index,
			cp.view_id,
			cp.score,
			p.title
		`,
		).
		Joins("JOIN problem AS p ON cp.problem_id = p.id").
		Where("cp.id = ?", contestId).
		Scan(&results).Error

	if err != nil {
		return nil, err
	}
	return results, nil
}

func (d *ContestProblemDao) GetProblemIdByContest(ctx context.Context, id int, index int) (*int, error) {
	var problemId int
	err := d.db.WithContext(ctx).
		Model(&foundationmodel.ContestProblem{}).
		Where("id = ? AND `index` = ?", id, index).
		Pluck("problem_id", &problemId).Error // index 是保留字，建议加反引号
	if err != nil {
		return nil, err
	}
	if problemId == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &problemId, nil
}
