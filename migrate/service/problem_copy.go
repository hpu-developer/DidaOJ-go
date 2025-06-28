package service

import (
	"context"
	"log/slog"
	metamysql "meta/meta-mysql"
	metatime "meta/meta-time"
	"strconv"
	"time"

	foundationdao "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model-mongo"
	metaerror "meta/meta-error"
	"meta/singleton"
)

type MigrateProblemCopyService struct {
}

var singletonMigrateProblemCopyService = singleton.Singleton[MigrateProblemCopyService]{}

func GetMigrateProblemCopyService() *MigrateProblemCopyService {
	return singletonMigrateProblemCopyService.GetInstance(
		func() *MigrateProblemCopyService {
			return &MigrateProblemCopyService{}
		},
	)
}

// GORM 模型定义
type JolProblem struct {
	ProblemID    int       `gorm:"column:problem_id"`
	Title        string    `gorm:"column:title"`
	Description  string    `gorm:"column:description"`
	Input        string    `gorm:"column:input"`
	Output       string    `gorm:"column:output"`
	SampleInput  string    `gorm:"column:sample_input"`
	SampleOutput string    `gorm:"column:sample_output"`
	Hint         string    `gorm:"column:hint"`
	Source       string    `gorm:"column:source"`
	TimeLimit    int       `gorm:"column:time_limit"`
	MemoryLimit  int       `gorm:"column:memory_limit"`
	InsertTime   time.Time `gorm:"column:in_date"`
}

func (JolProblem) TableName() string {
	return "problem"
}

type JolProblemTag struct {
	ProblemID int    `gorm:"column:problem_id"`
	Name      string `gorm:"column:name"`
}

func (JolProblemTag) TableName() string {
	return "problem_tag"
}

func (s *MigrateProblemCopyService) Start() error {
	ctx := context.Background()

	// 初始化 GORM 客户端
	codeojDB := metamysql.GetSubsystem().GetClient("jol")

	// 查询题目主表并构造 Mongo 对象
	var problems []JolProblem
	if err := codeojDB.Where("problem_id >= 2141").Find(&problems).Error; err != nil {
		return metaerror.Wrap(err, "query problems failed")
	}

	var problemDocs []*foundationmodel.Problem

	for _, p := range problems {
		seq, err := foundationdao.GetCounterDao().GetNextSequence(ctx, "problem_id")
		if err != nil {
			return err
		}
		newProblemId := strconv.Itoa(seq)

		description := p.Description
		if p.Input != "" {
			description += "\n\n## 输入\n" + p.Input
		}
		if p.Output != "" {
			description += "\n\n## 输出\n" + p.Output
		}
		if p.SampleInput != "" {
			description += "\n\n## 样例输入\n```\n" + p.SampleInput + "\n```\n"
		}
		if p.SampleOutput != "" {
			description += "\n\n## 样例输出\n```\n" + p.SampleOutput + "\n```\n"
		}
		if p.Hint != "" {
			description += "\n\n## 提示\n" + p.Hint
		}

		problemDocs = append(
			problemDocs, foundationmodel.NewProblemBuilder().
				Id(newProblemId).
				Sort(len(newProblemId)).
				Title(p.Title).
				Description(description).
				Source(&p.Source).
				TimeLimit(p.TimeLimit*1000).
				MemoryLimit(p.MemoryLimit*1024).
				Accept(0).
				Attempt(0).
				InsertTime(p.InsertTime).
				UpdateTime(metatime.GetTimeNow()).
				Build(),
		)
	}

	//// 插入 MongoDB
	if len(problemDocs) > 0 {
		err := foundationdao.GetProblemDao().UpdateProblemsExcludeManualEdit(ctx, problemDocs)
		if err != nil {
			return err
		}
		slog.Info("update problem success")
	}

	slog.Info("migrate problem success")

	return nil
}
