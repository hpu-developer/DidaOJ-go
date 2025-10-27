package service

import (
	"context"
	"fmt"
	foundationdao "foundation/foundation-dao-mongo"
	foundationmodel "foundation/foundation-model-mongo"
	foundationservice "foundation/foundation-service"
	"log/slog"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	"migrate/config"
	migratedao "migrate/dao"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type MigrateProblemDmojService struct {
}

// GORM 模型定义
type DmojProblem struct {
	Id          int       `gorm:"column:id"`
	ProblemId   string    `gorm:"column:problem_id"`
	Title       string    `gorm:"column:title"`
	TimeLimit   int       `gorm:"column:time_limit"`
	MemoryLimit int       `gorm:"column:memory_limit"`
	Description string    `gorm:"column:description"`
	Input       string    `gorm:"column:input"`
	Output      string    `gorm:"column:output"`
	Examples    string    `gorm:"column:examples"`
	Source      string    `gorm:"column:source"`
	Hint        string    `gorm:"column:hint"`
	GmtCreate   time.Time `gorm:"column:gmt_create"`
	GmtModified time.Time `gorm:"column:gmt_modified"`
}

func (DmojProblem) TableName() string {
	return "problem"
}

var singletonMigrateProblemDmojService = singleton.Singleton[MigrateProblemDmojService]{}

func GetMigrateProblemDmojService() *MigrateProblemDmojService {
	return singletonMigrateProblemDmojService.GetInstance(
		func() *MigrateProblemDmojService {
			return &MigrateProblemDmojService{}
		},
	)
}

func (s *MigrateProblemDmojService) Start() error {
	ctx := context.Background()

	// 初始化 GORM 客户端
	dmojDb := metapostgresql.GetSubsystem().GetClient("dmoj")

	var problemModels []DmojProblem
	if err := dmojDb.
		Model(&DmojProblem{}).
		Where("description IS NOT NULL").
		Where("description <> ''").
		Find(&problemModels).Error; err != nil {
		return metaerror.Wrap(err, "query dmoj problem failed")
	}

	for _, problemModel := range problemModels {

		description := "## 题目描述\n\n" + problemModel.Description

		if problemModel.Input != "" {
			description += "\n\n## 输入\n\n" + problemModel.Input
		}
		if problemModel.Output != "" {
			description += "\n\n## 输出\n\n" + problemModel.Output
		}
		sample := ""
		if problemModel.Examples != "" {
			re := regexp.MustCompile(`<input>(.*?)</input><output>(.*?)</output>`)
			matches := re.FindAllStringSubmatch(problemModel.Examples, -1)
			for i, match := range matches {
				sample += fmt.Sprintf("\n\n### 输入 #%d", i+1) + "\n\n```\n" + match[1] + "\n```\n"
				sample += fmt.Sprintf("\n\n### 输出 #%d", i+1) + "\n\n```\n" + match[2] + "\n```\n"
			}
			if len(matches) <= 1 {
				sample = strings.Replace(sample, "### 输入 #1", "## 样例输入", 1)
				sample = strings.Replace(sample, "### 输出 #1", "## 样例输出", 1)
			} else {
				sample = "\n\n## 样例\n\n" + sample
			}
		}
		if sample != "" {
			description += sample
		}
		if problemModel.Hint != "" {
			description += "\n\n## 提示\n\n" + problemModel.Hint
		}

		problem := foundationmodel.NewProblemBuilder().
			Title(problemModel.Title).
			Description(description).
			Source(&problemModel.Source).
			TimeLimit(problemModel.TimeLimit).
			MemoryLimit(problemModel.MemoryLimit * 1024).
			InsertTime(problemModel.GmtCreate).
			UpdateTime(problemModel.GmtModified).
			CreatorId(1650).
			Build()

		slog.Info("migrate dmoj problem", "id", problemModel.Id, "problem", problem)

		var tags []string
		if strings.HasPrefix(problemModel.ProblemId, "A") || strings.HasPrefix(problemModel.ProblemId, "B") {
			tags = append(tags, "离散数学")
		}
		if strings.HasPrefix(problemModel.ProblemId, "C") {
			tags = append(tags, "图论")
		}

		newId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "dmoj-problem", strconv.Itoa(problemModel.Id))
		if err != nil {
			return metaerror.Wrap(err, "get dmoj problem mark failed")
		}
		if newId != nil {
			err := foundationdao.GetProblemDao().UpdateProblem(ctx, *newId, problem, tags)
			if err != nil {
				return err
			}
			judgeDataRootPath := "C:\\Users\\BoilT\\OneDrive\\Backup\\ServerBackup\\DidaOJ\\DM\\testcase"

			md5, err := foundationdao.GetProblemDao().GetProblemJudgeMd5(ctx, *newId)
			if err != nil {
				return err
			}
			judgeDataPath := path.Join(judgeDataRootPath, "problem_"+strconv.Itoa(problemModel.Id))
			err = foundationservice.GetProblemService().PostJudgeData(
				ctx,
				*newId,
				judgeDataPath,
				md5,
				config.GetConfig().GoJudge.Url,
				nil,
				true,
			)
			if err != nil {
				slog.Error("upload judge data failed", "id", problemModel.Id, "newId", newId, "error", err)
				return err
			}
		} else {
			problemId, err := foundationdao.GetProblemDao().PostCreate(ctx, problem, tags)
			if err != nil {
				return err
			}
			err = migratedao.GetMigrateMarkDao().Mark(ctx, "dmoj-problem", strconv.Itoa(problemModel.Id), *problemId)
			if err != nil {
				return metaerror.Wrap(err, "mark dmoj problem failed")
			}
		}

	}

	slog.Info("ces", "len", len(problemModels))

	return nil
}
