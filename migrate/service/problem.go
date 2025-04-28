package service

import (
	"context"
	"database/sql"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	metapanic "meta/meta-panic"
	"meta/singleton"
	"strconv"
	"time"
)

type MigrateProblemService struct{}

var singletonMigrateProblemService = singleton.Singleton[MigrateProblemService]{}

func GetMigrateProblemService() *MigrateProblemService {
	return singletonMigrateProblemService.GetInstance(
		func() *MigrateProblemService {
			return &MigrateProblemService{}
		},
	)
}

func (s *MigrateProblemService) Start() error {

	ctx := context.Background()

	mysqlClient := metamysql.GetSubsystem().GetClient()

	// Problem 定义
	type Problem struct {
		ProblemID   int
		Title       sql.NullString
		Description sql.NullString
		Hint        sql.NullString
		Source      sql.NullString
		Creator     sql.NullString
		Privilege   sql.NullInt64
		TimeLimit   sql.NullInt64
		MemoryLimit sql.NullInt64
		JudgeType   sql.NullInt64
		Accept      sql.NullInt64
		Attempt     sql.NullInt64
		InsertTime  sql.NullTime
		UpdateTime  sql.NullTime
	}
	type Tag struct {
		Name string
	}
	rows, err := mysqlClient.Query("SELECT DISTINCT name FROM problem_tag WHERE name IS NOT NULL")
	if err != nil {
		return metaerror.Wrap(err, "query problem_tag failed")
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(rows)

	tagMap := make(map[string]int) // name -> key
	tagKey := 1

	var mongoTags []*foundationmodel.ProblemTag

	for rows.Next() {
		var tag Tag
		if err := rows.Scan(&tag.Name); err != nil {
			return metaerror.Wrap(err, "query problem_tag row failed")
		}
		tagMap[tag.Name] = tagKey
		mongoTags = append(mongoTags,
			foundationmodel.NewProblemTagBuilder().
				Id(strconv.Itoa(tagKey)).Name(tag.Name).
				Build(),
		)
		tagKey++
	}

	if len(mongoTags) > 0 {
		err = foundationdao.GetProblemTagDao().UpdateProblemTags(ctx, mongoTags)
		if err != nil {
			return err
		}
		slog.Info("update problem tags success")
	}

	// === 拉取 problem 表 ===
	problemRows, err := mysqlClient.Query(`
		SELECT problem_id, title, description, hint, source, creator, privilege,
		       time_limit, memory_limit, judge_type, accept, attempt, insert_time, update_time
		FROM problem
	`)
	if err != nil {
		return metaerror.Wrap(err, "query problem row failed")
	}
	defer func(problemRows *sql.Rows) {
		err := problemRows.Close()
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(problemRows)

	// 先拉取所有 problem_tag 映射 (problem_id -> []tagName)
	problemTagMap := make(map[int][]int) // problem_id -> []tagKey
	tagRelRows, err := mysqlClient.Query(`SELECT problem_id, name FROM problem_tag WHERE name IS NOT NULL`)
	if err != nil {
		return metaerror.Wrap(err, "query problem_tag failed")
	}
	defer func(tagRelRows *sql.Rows) {
		err := tagRelRows.Close()
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(tagRelRows)

	for tagRelRows.Next() {
		var pid int
		var name string
		if err := tagRelRows.Scan(&pid, &name); err != nil {
			return metaerror.Wrap(err, "query problem_tag row failed")
		}
		key, ok := tagMap[name]
		if ok {
			problemTagMap[pid] = append(problemTagMap[pid], key)
		}
	}

	// === 处理每一条 problem 并插入 MongoDB ===
	var problemDocs []*foundationmodel.Problem

	for problemRows.Next() {
		var p Problem
		if err := problemRows.Scan(
			&p.ProblemID, &p.Title, &p.Description, &p.Hint, &p.Source, &p.Creator, &p.Privilege,
			&p.TimeLimit, &p.MemoryLimit, &p.JudgeType, &p.Accept, &p.Attempt, &p.InsertTime, &p.UpdateTime,
		); err != nil {
			return metaerror.Wrap(err, "query problem row failed")
		}

		seq, err := foundationdao.GetCounterDao().GetNextSequence(ctx, "problem_id")
		if err != nil {
			return err
		}

		problemDocs = append(problemDocs, foundationmodel.NewProblemBuilder().
			Id(strconv.Itoa(seq)).
			Title(nullStringToString(p.Title)).
			Description(nullStringToString(p.Description)).
			Hint(nullStringToString(p.Hint)).
			Source(nullStringToString(p.Source)).
			Creator(nullStringToString(p.Creator)).
			Privilege(int(p.Privilege.Int64)).
			TimeLimit(int(p.TimeLimit.Int64)*1000).
			MemoryLimit(int(p.MemoryLimit.Int64)*1024).
			JudgeType(int(p.JudgeType.Int64)).
			Tags(problemTagMap[p.ProblemID]).
			Accept(int(p.Accept.Int64)).
			Attempt(int(p.Attempt.Int64)).
			InsertTime(nullTimeToTime(p.InsertTime)).
			UpdateTime(nullTimeToTime(p.UpdateTime)).
			Build())
	}

	// 插入 MongoDB
	if len(problemDocs) > 0 {
		//err = problemCol.Drop(ctx) // 清空原 problem 集合
		//if err != nil {
		//	log.Fatal("清空 problem 出错:", err)
		//}

		err = foundationdao.GetProblemDao().UpdateProblems(ctx, problemDocs)
		if err != nil {
			return err
		}
		slog.Info("update problem success")
	}

	slog.Info("migrate problem success")

	return nil
}

func nullStringToString(s sql.NullString) string {
	if s.Valid {
		return s.String
	}
	return ""
}

func nullTimeToTime(t sql.NullTime) time.Time {
	if t.Valid {
		return t.Time
	}
	return time.Time{}
}
