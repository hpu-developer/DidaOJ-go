package service

import (
	"context"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	foundationservice "foundation/foundation-service"
	"log/slog"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	metapath "meta/meta-path"
	metastring "meta/meta-string"
	"meta/singleton"
	"migrate/config"
	migratedao "migrate/dao"
	"path"
	"slices"
	"strconv"
	"strings"
	"time"
)

type MigrateProblemEojService struct {
}

// GORM 模型定义
type EojProblem struct {
	Id          int       `gorm:"column:id"`
	Title       string    `gorm:"column:title"`
	Description string    `gorm:"column:description"`
	Input       string    `gorm:"column:input"`
	Output      string    `gorm:"column:output"`
	Sample      string    `gorm:"column:sample"`
	Hint        string    `gorm:"column:hint"`
	Source      string    `gorm:"column:source"`
	CreateTime  time.Time `gorm:"column:create_time"`
	UpdateTime  time.Time `gorm:"column:update_time"`
	TimeLimit   int       `gorm:"column:time_limit"`
	MemoryLimit int       `gorm:"column:memory_limit"`
	Cases       string    `gorm:"column:cases"`
}

func (EojProblem) TableName() string {
	return "problem_problem"
}

type EojProblemManager struct {
	Id        int `gorm:"column:id"`
	ProblemId int `gorm:"column:problem_id"`
	UserId    int `gorm:"column:user_id"`
}

func (EojProblemManager) TableName() string {
	return "problem_problem_managers"
}

type EojProblemTaggedItem struct {
	Id       int `gorm:"column:id"`
	ObjectId int `gorm:"column:object_id"`
	TagId    int `gorm:"column:tag_id"`
}

func (EojProblemTaggedItem) TableName() string {
	return "tagging_taggeditem"
}

type EojProblemTag struct {
	Id   int    `gorm:"column:id"`
	Name string `gorm:"column:name"`
}

func (EojProblemTag) TableName() string {
	return "tagging_tag"
}

var singletonMigrateProblemV2Service = singleton.Singleton[MigrateProblemEojService]{}

func GetMigrateProblemEojService() *MigrateProblemEojService {
	return singletonMigrateProblemV2Service.GetInstance(
		func() *MigrateProblemEojService {
			return &MigrateProblemEojService{}
		},
	)
}

func (s *MigrateProblemEojService) Start() error {
	ctx := context.Background()

	ignoreProblem := []int{
		30, 49, 66, 67, 425, 439, 440, 441, 443,
	}
	migrateProblem := map[int]string{
		1:   "134",
		50:  "455",
		51:  "456",
		52:  "457",
		53:  "458",
		54:  "459",
		55:  "460",
		56:  "461",
		57:  "462",
		58:  "463",
		59:  "464",
		60:  "465",
		61:  "466",
		62:  "467",
		298: "43",
		299: "101",
		301: "101",
		302: "102",
		303: "103",
		304: "104",
		305: "105",
		306: "106",
		307: "107",
		308: "108",
		309: "109",
		310: "110",
		311: "111",
		312: "112",
		313: "113",
		314: "114",
		315: "115",
		316: "116",
		317: "117",
		318: "118",
		319: "119",
		320: "120",
		321: "121",
		322: "122",
		323: "123",
		324: "124",
		325: "125",
		326: "126",
		327: "127",
		328: "128",
		329: "129",
		330: "130",
		331: "131",
		376: "134",
		407: "134",
	}
	// 初始化 GORM 客户端
	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	var problemModels []EojProblem
	if err := eojDb.
		Model(&EojProblem{}).
		Where("description IS NOT NULL").
		Where("description <> ''").
		Find(&problemModels).Error; err != nil {
		return metaerror.Wrap(err, "query eoj problem failed")
	}

	for _, problemModel := range problemModels {
		migrateId, ok := migrateProblem[problemModel.Id]
		if ok {
			err := migratedao.GetMigrateMarkDao().Mark(ctx, "eoj-problem", strconv.Itoa(problemModel.Id), migrateId)
			if err != nil {
				return metaerror.Wrap(err, "mark eoj problem failed")
			}
			continue
		}
		if slices.Contains(ignoreProblem, problemModel.Id) {
			slog.Info("ignore eoj problem", "id", problemModel.Id, "title", problemModel.Title)
			continue
		}
		var problemModelManagers []EojProblemManager
		if err := eojDb.
			Model(&EojProblemManager{}).
			Where("problem_id = ?", problemModel.Id).
			Find(&problemModelManagers).Error; err != nil {
			return metaerror.Wrap(err, "query eoj problem managers failed")
		}

		userId := 3441
		var authMembers []int
		if len(problemModelManagers) > 0 {
			userId = problemModelManagers[0].UserId
			realId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(userId))
			if err != nil {
				return metaerror.Wrap(err, "get eoj user mark failed")
			}
			if realId == nil {
				return metaerror.Wrap(err, "get eoj user mark failed")
			}
			userId, err = strconv.Atoi(*realId)
			if err != nil {
				return metaerror.Wrap(err, "convert eoj user mark to int failed")
			}
			for _, manager := range problemModelManagers {
				realId, err = migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(manager.UserId))
				if err != nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				if realId == nil {
					return metaerror.Wrap(err, "get eoj user mark failed")
				}
				authUserId, err := strconv.Atoi(*realId)
				if err != nil {
					return metaerror.Wrap(err, "convert eoj user mark to int failed")
				}
				authMembers = append(authMembers, authUserId)
			}
		}

		description := "## 题目描述\n\n" + problemModel.Description

		if problemModel.Input != "" {
			description += "\n\n## 输入\n\n" + problemModel.Input
		}
		if problemModel.Output != "" {
			description += "\n\n## 输出\n\n" + problemModel.Output
		}
		sample := ""
		if problemModel.Sample != "" {
			sampleSlices := strings.Split(problemModel.Sample, ",")
			index := 1
			for _, sampleSlice := range sampleSlices {
				fileA := sampleSlice[0:2]
				fileB := sampleSlice[2:4]
				fileRoot := "C:\\Users\\BoilT\\OneDrive\\Backup\\ServerBackup\\DidaOJ\\eoj\\eoj\\eoj\\data\\"
				filePathIn := path.Join(
					fileRoot,
					"in",
					fileA,
					fileB,
					sampleSlice,
				)
				filePathOut := path.Join(
					fileRoot,
					"out",
					fileA,
					fileB,
					sampleSlice,
				)
				inString, _ := metastring.GetStringFromOpenFile(filePathIn)
				outString, _ := metastring.GetStringFromOpenFile(filePathOut)
				if inString != "" || outString != "" {
					if inString != "" {
						sample += fmt.Sprintf("\n\n### 输入 #%d", index) + "\n\n```\n" + inString + "\n```\n"
					}
					if outString != "" {
						sample += fmt.Sprintf("\n\n### 输出 #%d", index) + "\n\n```\n" + outString + "\n```\n"
					}
					index++
				} else {
					slog.Warn("empty sample", "file", sampleSlice, "in", inString, "out", outString)
				}
			}
			if index <= 2 {
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
			Source(problemModel.Source).
			TimeLimit(problemModel.TimeLimit).
			MemoryLimit(problemModel.MemoryLimit * 1024).
			InsertTime(problemModel.CreateTime).
			UpdateTime(problemModel.UpdateTime).
			CreatorId(userId).
			AuthMembers(authMembers).
			Build()

		judgeDataRootPath := "C:\\Users\\BoilT\\OneDrive\\Backup\\ServerBackup\\DidaOJ\\eoj\\judge_data"
		casesSlices := strings.Split(problemModel.Cases, ",")
		index := 1
		for _, casesSlice := range casesSlices {
			fileA := casesSlice[0:2]
			fileB := casesSlice[2:4]
			fileRoot := "C:\\Users\\BoilT\\OneDrive\\Backup\\ServerBackup\\DidaOJ\\eoj\\eoj\\eoj\\data\\"
			filePathIn := path.Join(
				fileRoot,
				"in",
				fileA,
				fileB,
				casesSlice,
			)
			filePathOut := path.Join(
				fileRoot,
				"out",
				fileA,
				fileB,
				casesSlice,
			)
			inString, _ := metastring.GetStringFromOpenFile(filePathIn)
			outString, _ := metastring.GetStringFromOpenFile(filePathOut)
			if inString != "" || outString != "" {
				key := fmt.Sprintf("%02d", index)
				if inString != "" {
					savePath := path.Join(judgeDataRootPath, strconv.Itoa(problemModel.Id), key+".in")
					err := metapath.WriteStringToFile(savePath, inString)
					if err != nil {
						return metaerror.Wrap(err, "write eoj problem input file failed")
					}
				}
				if outString != "" {
					savePath := path.Join(judgeDataRootPath, strconv.Itoa(problemModel.Id), key+".out")
					err := metapath.WriteStringToFile(savePath, outString)
					if err != nil {
						return metaerror.Wrap(err, "write eoj problem output file failed")
					}
				}
				index++
			} else {
				slog.Warn(
					"empty test case",
					"id",
					problemModel.Id,
					"file",
					casesSlice,
					"in",
					inString,
					"out",
					outString,
				)
			}
		}

		var problemTags []EojProblemTaggedItem
		if err := eojDb.
			Model(&EojProblemTaggedItem{}).
			Where("object_id = ?", problemModel.Id).
			Find(&problemTags).Error; err != nil {
			return metaerror.Wrap(err, "query eoj problem managers failed")
		}

		var tags []string

		for _, problemTag := range problemTags {
			var tagModel EojProblemTag
			if err := eojDb.
				Model(&EojProblemTag{}).
				Where("id = ?", problemTag.TagId).
				Find(&tagModel).Error; err != nil {
				return metaerror.Wrap(err, "query eoj problem managers failed")
			}
			tags = append(tags, tagModel.Name)
		}

		slog.Info("migrate eoj problem", "id", problemModel.Id, "problem", problem)

		newId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-problem", strconv.Itoa(problemModel.Id))
		if err != nil {
			return metaerror.Wrap(err, "get eoj problem mark failed")
		}
		if newId != nil {
			err := foundationdao.GetProblemDao().UpdateProblem(ctx, *newId, problem, tags)
			if err != nil {
				return err
			}
		} else {
			problemId, err := foundationdao.GetProblemDao().PostCreate(ctx, problem, tags)
			if err != nil {
				return err
			}
			err = migratedao.GetMigrateMarkDao().Mark(ctx, "eoj-problem", strconv.Itoa(problemModel.Id), *problemId)
			if err != nil {
				return metaerror.Wrap(err, "mark eoj problem failed")
			}
			newId = problemId
		}

		uploadJudgeData := true
		if uploadJudgeData {
			md5, err := foundationdao.GetProblemDao().GetProblemJudgeMd5(ctx, *newId)
			if err != nil {
				return err
			}
			judgeDataPath := path.Join(judgeDataRootPath, strconv.Itoa(problemModel.Id))
			err = foundationservice.GetProblemService().PostJudgeData(
				ctx,
				*newId,
				path.Join(judgeDataPath),
				md5,
				config.GetConfig().GoJudge.Url,
			)
			if err != nil {
				slog.Error("upload judge data failed", "id", problemModel.Id, "newId", newId, "error", err)
				return err
			}
		}
	}

	slog.Info("ces", "len", len(problemModels))

	return nil
}
