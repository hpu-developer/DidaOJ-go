package service

import (
	"context"
	"log/slog"
	metamysql "meta/meta-mysql"
	"strconv"
	"time"

	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	"meta/singleton"
)

type MigrateProblemService struct {
	oldProblemIdToNewProblemId map[int]string
}

var singletonMigrateProblemService = singleton.Singleton[MigrateProblemService]{}

func GetMigrateProblemService() *MigrateProblemService {
	return singletonMigrateProblemService.GetInstance(
		func() *MigrateProblemService {
			return &MigrateProblemService{}
		},
	)
}

// GORM 模型定义
type Problem struct {
	ProblemID   int       `gorm:"column:problem_id"`
	Title       string    `gorm:"column:title"`
	Description string    `gorm:"column:description"`
	Hint        string    `gorm:"column:hint"`
	Source      string    `gorm:"column:source"`
	Creator     string    `gorm:"column:creator"`
	Privilege   int       `gorm:"column:privilege"`
	TimeLimit   int       `gorm:"column:time_limit"`
	MemoryLimit int       `gorm:"column:memory_limit"`
	JudgeType   int       `gorm:"column:judge_type"`
	Accept      int       `gorm:"column:accept"`
	Attempt     int       `gorm:"column:attempt"`
	InsertTime  time.Time `gorm:"column:insert_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
}

type ProblemTag struct {
	ProblemID int    `gorm:"column:problem_id"`
	Name      string `gorm:"column:name"`
}

func (Problem) TableName() string {
	return "problem"
}

func (ProblemTag) TableName() string {
	return "problem_tag"
}

func (s *MigrateProblemService) Start() error {
	ctx := context.Background()

	// 初始化 GORM 客户端
	codeojDB := metamysql.GetSubsystem().GetClient("codeoj")

	// 查询所有唯一标签
	var tags []ProblemTag
	if err := codeojDB.
		Model(&ProblemTag{}).
		Select("DISTINCT name").
		Where("name IS NOT NULL").
		Scan(&tags).Error; err != nil {
		return metaerror.Wrap(err, "query problem_tag failed")
	}

	tagMap := make(map[string]int)
	var mongoTags []*foundationmodel.ProblemTag
	tagKey := 1

	for _, tag := range tags {
		tagMap[tag.Name] = tagKey
		mongoTags = append(mongoTags, foundationmodel.NewProblemTagBuilder().Id(tagKey).Name(tag.Name).Build())
		tagKey++
	}

	if len(mongoTags) > 0 {
		err := foundationdao.GetProblemTagDao().UpdateProblemTags(ctx, mongoTags)
		if err != nil {
			return err
		}
		slog.Info("update problem tags success")
	}

	// 查询题目和标签关系表
	var tagModels []ProblemTag
	if err := codeojDB.
		Model(&ProblemTag{}).
		Where("name IS NOT NULL").
		Find(&tagModels).Error; err != nil {
		return metaerror.Wrap(err, "query problem_tag rels failed")
	}

	problemTagMap := map[int][]int{}
	for _, rel := range tagModels {
		if key, ok := tagMap[rel.Name]; ok {
			problemTagMap[rel.ProblemID] = append(problemTagMap[rel.ProblemID], key)
		}
	}

	// 查询题目主表并构造 Mongo 对象
	var problems []Problem
	if err := codeojDB.Find(&problems).Error; err != nil {
		return metaerror.Wrap(err, "query problems failed")
	}

	var problemDocs []*foundationmodel.Problem
	s.oldProblemIdToNewProblemId = make(map[int]string)
	s.oldProblemIdToNewProblemId[0] = "1000"

	for _, p := range problems {
		seq, err := foundationdao.GetCounterDao().GetNextSequence(ctx, "problem_id")
		if err != nil {
			return err
		}
		newProblemId := strconv.Itoa(seq)
		s.oldProblemIdToNewProblemId[p.ProblemID] = newProblemId

		description := p.Description
		if p.Hint != "" {
			description += "\n\n## 提示\n" + p.Hint
		}

		problemDocs = append(problemDocs, foundationmodel.NewProblemBuilder().
			Id(newProblemId).
			Sort(len(newProblemId)).
			Title(p.Title).
			Description(description).
			Source(p.Source).
			Creator(p.Creator).
			Privilege(p.Privilege).
			TimeLimit(p.TimeLimit*1000).
			MemoryLimit(p.MemoryLimit*1024).
			JudgeType(foundationmodel.JudgeType(p.JudgeType)).
			Tags(problemTagMap[p.ProblemID]).
			Accept(p.Accept).
			Attempt(p.Attempt).
			InsertTime(p.InsertTime).
			UpdateTime(p.UpdateTime).
			Build())
	}

	// 插入 MongoDB
	if len(problemDocs) > 0 {
		err := foundationdao.GetProblemDao().UpdateProblems(ctx, problemDocs)
		if err != nil {
			return err
		}
		slog.Info("update problem success")
	}

	slog.Info("migrate problem success")

	return nil
}

func (s *MigrateProblemService) GetNewProblemId(oldProblemId int) string {
	if id, ok := s.oldProblemIdToNewProblemId[oldProblemId]; ok {
		return id
	}
	return strconv.Itoa(oldProblemId)
}
