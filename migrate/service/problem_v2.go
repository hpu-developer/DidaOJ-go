package service

import (
	"context"
	"log/slog"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	migratedao "migrate/dao"
	"slices"
	"strconv"
	"time"
)

type MigrateProblemV2Service struct {
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

var singletonMigrateProblemV2Service = singleton.Singleton[MigrateProblemV2Service]{}

func GetMigrateProblemV2Service() *MigrateProblemV2Service {
	return singletonMigrateProblemV2Service.GetInstance(
		func() *MigrateProblemV2Service {
			return &MigrateProblemV2Service{}
		},
	)
}

func (s *MigrateProblemV2Service) Start() error {
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
		return metaerror.Wrap(err, "query problem_tag rels failed")
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

		//description := problemModel.Description
		//
		//problem := foundationmodel.NewProblemBuilder().
		//	Title(problemModel.Title).
		//	Description(description).
		//	Source(problemModel.Source).
		//	TimeLimit(problemModel.TimeLimit).
		//	MemoryLimit(problemModel.MemoryLimit * 1024).
		//	InsertTime(problemModel.CreateTime).
		//	UpdateTime(problemModel.UpdateTime).
		//	CreatorId(userId).
		//	Build()
		//
		//problemModel, err := foundationdao.GetProblemDao().PostCreate(ctx, problem, tags)
		//if err != nil {
		//	return err
		//}
		//
		//err := migratedao.GetMigrateMarkDao().MarkEojProblem(nil, problemModel.Id, newId)
		//if err != nil {
		//	return metaerror.Wrap(err, "mark eoj problem failed")
		//}
	}

	slog.Info("ces", "len", len(problemModels))

	return nil
}
