package service

import (
	"context"
	"fmt"
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
type JolProblem struct {
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

type VhojProblem struct {
	Id          int       `gorm:"column:C_ID"`
	OriginOj    string    `gorm:"column:C_originOJ"`
	OriginProb  string    `gorm:"column:C_originProb"`
	Title       string    `gorm:"column:C_TITLE"`
	Source      string    `gorm:"column:C_SOURCE"`
	Url         string    `gorm:"column:C_URL"`
	MemoryLimit int       `gorm:"column:C_MEMORYLIMIT"`
	TimeLimit   int       `gorm:"column:C_TIMELIMIT"`
	TriggerTime time.Time `gorm:"column:C_TRIGGER_TIME"`
}

func (VhojProblem) TableName() string {
	return "t_problem"
}

type VhojDescription struct {
	Id           int       `gorm:"column:C_ID"`
	ProblemId    int       `gorm:"column:C_PROBLEM_ID"`
	UpdateTime   time.Time `gorm:"column:C_UPDATE_TIME"`
	Description  string    `gorm:"column:C_DESCRIPTION"`
	Input        string    `gorm:"column:C_INPUT"`
	Output       string    `gorm:"column:C_OUTPUT"`
	SampleInput  string    `gorm:"column:C_SAMPLEINPUT"`
	SampleOutput string    `gorm:"column:C_SAMPLEOUTPUT"`
	Hint         string    `gorm:"column:C_HINT"`
	Author       string    `gorm:"column:C_AUTHOR"`
}

func (VhojDescription) TableName() string {
	return "t_description"
}

func (s *MigrateProblemService) Start() error {
	ctx := context.Background()

	err := s.processCodeOjProblem(ctx)
	if err != nil {
		return err
	}

	err = s.processVhojProblem(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (s *MigrateProblemService) processCodeOjProblem(ctx context.Context) error {

	// 初始化 GORM 客户端
	codeojDB := metamysql.GetSubsystem().GetClient("codeoj")

	// 查询所有唯一标签
	var tags []JolProblemTag
	if err := codeojDB.
		Model(&JolProblemTag{}).
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
	var tagModels []JolProblemTag
	if err := codeojDB.
		Model(&JolProblemTag{}).
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
	var problems []JolProblem
	if err := codeojDB.Find(&problems).Error; err != nil {
		return metaerror.Wrap(err, "query problems failed")
	}

	var problemDocs []*foundationmodel.Problem
	s.oldProblemIdToNewProblemId = make(map[int]string)
	s.oldProblemIdToNewProblemId[1000] = "1"

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
			CreatorNickname(p.Creator).
			Privilege(p.Privilege).
			TimeLimit(p.TimeLimit*1000).
			MemoryLimit(p.MemoryLimit*1024).
			JudgeType(foundationmodel.JudgeType(p.JudgeType)).
			Tags(problemTagMap[p.ProblemID]).
			Accept(0).
			Attempt(0).
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

func (s *MigrateProblemService) processVhojProblem(ctx context.Context) error {
	vhojDB := metamysql.GetSubsystem().GetClient("vhoj")

	var problems []VhojProblem
	if err := vhojDB.Find(&problems).Error; err != nil {
		return metaerror.Wrap(err, "query problems failed")
	}

	var problemDocs []*foundationmodel.Problem

	for _, p := range problems {
		if p.OriginOj == "HPU" {
			continue
		}
		newProblemId := fmt.Sprintf("%s-%s", p.OriginOj, p.OriginProb)

		var vhojDescription VhojDescription
		if err := vhojDB.Where("C_PROBLEM_ID = ?", p.Id).First(&vhojDescription).Error; err != nil {
			return metaerror.Wrap(err, "query problem description failed")
		}

		description := ""

		if vhojDescription.Description != "" {
			description += "## Description\n" + vhojDescription.Description
		}

		if vhojDescription.Input != "" {
			if description != "" {
				description += "\n\n"
			}
			description += "## Input\n" + vhojDescription.Input
		}

		if vhojDescription.Output != "" {
			if description != "" {
				description += "\n\n"
			}
			description += "## Output\n" + vhojDescription.Output
		}

		if vhojDescription.SampleInput != "" {
			if description != "" {
				description += "\n\n"
			}
			description += "## Sample Input\n" + vhojDescription.SampleInput
		}

		if vhojDescription.SampleOutput != "" {
			if description != "" {
				description += "\n\n"
			}
			description += "## Sample Output\n" + vhojDescription.SampleOutput
		}
		if vhojDescription.Hint != "" {
			if description != "" {
				description += "\n\n"
			}
			description += "## Hint\n" + vhojDescription.Hint
		}

		problemDocs = append(problemDocs, foundationmodel.NewProblemBuilder().
			Id(newProblemId).
			OriginOj(p.OriginOj).
			OriginId(p.OriginProb).
			OriginUrl(p.Url).
			Sort(len(newProblemId)).
			Title(p.Title).
			Description(description).
			Source(p.Source).
			CreatorNickname(vhojDescription.Author).
			TimeLimit(p.TimeLimit).
			MemoryLimit(p.MemoryLimit).
			JudgeType(foundationmodel.JudgeTypeNormal).
			Accept(0).
			Attempt(0).
			InsertTime(p.TriggerTime).
			UpdateTime(vhojDescription.UpdateTime).
			Build())
	}

	// 插入 MongoDB
	if len(problemDocs) > 0 {
		err := foundationdao.GetProblemDao().UpdateProblems(ctx, problemDocs)
		if err != nil {
			return err
		}
	}

	slog.Info("migrate Vhoj problem success")

	return nil
}

func (s *MigrateProblemService) GetNewProblemId(oldProblemId int) string {
	if id, ok := s.oldProblemIdToNewProblemId[oldProblemId]; ok {
		return id
	}
	return "-1"
}
