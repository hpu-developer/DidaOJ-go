package foundationdao

import (
	"context"
	"gorm.io/gorm"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type JudgeJobCompileDao struct {
	db *gorm.DB
}

var singletonJudgeJobCompileDao = singleton.Singleton[JudgeJobCompileDao]{}

func GetJudgeJobCompileDao() *JudgeJobCompileDao {
	return singletonJudgeJobCompileDao.GetInstance(
		func() *JudgeJobCompileDao {
			dao := &JudgeJobCompileDao{}
			dao.db = metamysql.GetSubsystem().GetClient("didaoj")
			return dao
		},
	)
}

func (d *JudgeJobCompileDao) MarkJudgeJobCompileMessage(
	ctx context.Context,
	id int,
	judger string,
	message string,
) error {
	err := d.db.WithContext(ctx).
		Exec(
			`
		INSERT INTO judge_job_compile (id, message)
SELECT j.id, ?
FROM judge_job AS j
WHERE j.id = ?
  AND j.judger = ?
ON DUPLICATE KEY UPDATE message = ?;
	`, message, id, judger, message,
		).Error
	if err != nil {
		return metaerror.Wrap(err)
	}
	return nil
}
