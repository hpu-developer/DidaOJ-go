package foundationdao

import (
	"context"
	metaerror "meta/meta-error"
	metapostgresql "meta/meta-postgresql"
	metautf "meta/meta-utf"
	"meta/singleton"

	"gorm.io/gorm"
)

type JudgeJobCompileDao struct {
	db *gorm.DB
}

var singletonJudgeJobCompileDao = singleton.Singleton[JudgeJobCompileDao]{}

func GetJudgeJobCompileDao() *JudgeJobCompileDao {
	return singletonJudgeJobCompileDao.GetInstance(
		func() *JudgeJobCompileDao {
			dao := &JudgeJobCompileDao{}
			dao.db = metapostgresql.GetSubsystem().GetClient("didaoj")
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
	message = metautf.SanitizeText(message)
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
