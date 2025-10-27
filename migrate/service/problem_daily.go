package service

import (
	"context"
	"fmt"
	foundationdao "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model-mongo"
	"log/slog"
	metamysql "meta/meta-mysql"
	"strconv"
	"time"

	metaerror "meta/meta-error"
	"meta/singleton"
)

type MigrateProblemDailyService struct {
}

var singletonMigrateProblemDailyService = singleton.Singleton[MigrateProblemDailyService]{}

func GetMigrateProblemDailyService() *MigrateProblemDailyService {
	return singletonMigrateProblemDailyService.GetInstance(
		func() *MigrateProblemDailyService {
			return &MigrateProblemDailyService{}
		},
	)
}

// GORM 模型定义
type Daily struct {
	Time     time.Time `json:"time"`
	OJ       string    `gorm:"column:oj"`
	Id       string    `gorm:"column:id"`
	Title    string    `gorm:"column:title"`
	Content  string    `gorm:"column:content"`
	Solution string    `gorm:"column:solution"`
	Code     string    `gorm:"column:code"`
}

func (Daily) TableName() string {
	return "daily"
}

func (s *MigrateProblemDailyService) Start() error {
	//ctx := context.Background()

	// 初始化 GORM 客户端
	codeojDB := metapostgresql.GetSubsystem().GetClient("jol")

	// 查询题目主表并构造 Mongo 对象
	var problems []Daily
	if err := codeojDB.Find(&problems).Error; err != nil {
		return metaerror.Wrap(err, "query daily failed")
	}

	ctx := context.Background()

	createTime, err := time.Parse("2006-01-02 15:04:05", "2017-07-28 12:00:00")
	if err != nil {
		return metaerror.Wrap(err, "parse create time failed")
	}
	for _, problem := range problems {
		id := problem.Time.Format("2006-01-02")
		var problemId string
		if problem.OJ == "HPU" {
			number, _ := strconv.Atoi(problem.Id)
			problemId = strconv.Itoa(number - 999)
		} else {
			problemId = fmt.Sprintf("%s-%s", problem.OJ, problem.Id)
		}

		problemDoc := foundationmodel.NewProblemDailyBuilder().
			Id(id).
			ProblemId(problemId).
			Solution(problem.Solution).
			Code("```cpp\n" + problem.Code + "\n```").
			CreateTime(createTime).
			UpdateTime(createTime).
			CreatorId(3).
			UpdaterId(3).
			Build()
		err := foundationdao.GetProblemDailyDao().UpdateOrInsertProblemDaily(ctx, id, problemDoc)
		if err != nil {
			return err
		}
	}

	slog.Info("migrate daily success")

	return nil
}
