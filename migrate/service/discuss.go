package service

import (
	"context"
	"log/slog"
	"sort"
	"time"

	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
)

type MigrateDiscussService struct {
}

var singletonMigrateDiscussService = singleton.Singleton[MigrateDiscussService]{}

func GetMigrateDiscussService() *MigrateDiscussService {
	return singletonMigrateDiscussService.GetInstance(
		func() *MigrateDiscussService {
			s := &MigrateDiscussService{}
			return s
		},
	)
}

// GORM 模型定义
type JolTopic struct {
	Tid      int    `gorm:"column:tid"`
	Title    string `gorm:"column:title"`
	Cid      int    `gorm:"column:cid"`
	Pid      int    `gorm:"column:pid"`
	AuthorId string `gorm:"column:author_id"`
}

func (JolTopic) TableName() string {
	return "topic"
}

type JolReply struct {
	Rid      string    `gorm:"column:rid"`
	AuthorId string    `gorm:"column:author_id"`
	Content  string    `gorm:"column:content"`
	TopicId  int       `gorm:"column:topic_id"`
	Time     time.Time `gorm:"column:time"`
}

func (JolReply) TableName() string {
	return "reply"
}

type CodeojBlog struct {
	BlogId     string    `gorm:"column:blog_id"`
	Content    string    `gorm:"column:content"`
	Creator    string    `gorm:"column:creator"`
	PageViews  int       `gorm:"column:pageviews"`
	InsertTime time.Time `gorm:"column:insert_time"`
	UpdateTime time.Time `gorm:"column:update_time"`
	Title      string    `gorm:"column:title"`
}

func (CodeojBlog) TableName() string {
	return "blog"
}

func (s *MigrateDiscussService) Start() error {
	ctx := context.Background()

	var discusses []*foundationmodel.Discuss
	var jolReplyMap map[int][]JolReply
	jolDiscusses, jolReplyMap, err := s.processJolDiscuss(ctx)
	if err != nil {
		return err
	}
	discusses = append(discusses, jolDiscusses...)

	codeojBlogs, err := s.processCodeojBlog(ctx)
	if err != nil {
		return err
	}
	discusses = append(discusses, codeojBlogs...)

	slog.Info("migrate Discusses updates", "count", len(discusses))

	sort.Slice(discusses, func(i, j int) bool {
		return discusses[i].InsertTime.Before(discusses[j].InsertTime)
	})
	oldJolDiscussIdToNewDiscussId := make(map[int]int)
	for _, discuss := range discusses {
		var oldId int
		if discuss.Id > 0 {
			oldId = discuss.Id
		}
		err = foundationdao.GetDiscussDao().InsertDiscuss(ctx, discuss)
		if err != nil {
			return metaerror.Wrap(err, "insert Discuss failed")
		}
		if oldId > 0 {
			oldJolDiscussIdToNewDiscussId[oldId] = discuss.Id
		}
	}

	var jolComments []*foundationmodel.DiscussComment
	for _, replies := range jolReplyMap {
		for _, reply := range replies {
			userId, err := GetMigrateUserService().getUserIdByUsername(ctx, reply.AuthorId)
			if err != nil {
				return metaerror.Wrap(err, "get userId failed, userId: %s", reply.AuthorId)
			}
			newDiscussId, ok := oldJolDiscussIdToNewDiscussId[reply.TopicId]
			if !ok {
				return metaerror.New("oldJolDiscussIdToNewDiscussId not found, oldId: %d", reply.TopicId)
			}

			discussComment := foundationmodel.NewDiscussCommentBuilder().
				DiscussId(newDiscussId).
				Content(reply.Content).
				AuthorId(userId).
				InsertTime(reply.Time).
				UpdateTime(reply.Time).
				Build()
			jolComments = append(jolComments, discussComment)
		}
	}
	slog.Info("migrate DiscussReplies updates", "count", len(jolComments))

	sort.Slice(jolComments, func(i, j int) bool {
		return jolComments[i].InsertTime.Before(jolComments[j].InsertTime)
	})

	for _, reply := range jolComments {
		err = foundationdao.GetDiscussCommentDao().InsertDiscussComment(ctx, reply)
		if err != nil {
			return metaerror.Wrap(err, "insert DiscussComment failed")
		}
	}

	return nil
}

func (s *MigrateDiscussService) processJolDiscuss(ctx context.Context) ([]*foundationmodel.Discuss, map[int][]JolReply, error) {
	slog.Info("migrate Discuss processJolDiscuss")

	db := metamysql.GetSubsystem().GetClient("jol")

	var jolTopics []JolTopic
	if err := db.Order("tid ASC").Find(&jolTopics).Error; err != nil {
		return nil, nil, metaerror.Wrap(err, "query jolTopics failed")
	}
	var jolReplies []JolReply
	if err := db.Order("rid ASC").Find(&jolReplies).Error; err != nil {
		return nil, nil, metaerror.Wrap(err, "query jolReplies failed")
	}
	jolReplyMap := make(map[int][]JolReply)
	for _, u := range jolReplies {
		jolReplyMap[u.TopicId] = append(jolReplyMap[u.TopicId], u)
	}

	var docs []*foundationmodel.Discuss
	for _, u := range jolTopics {
		replies := jolReplyMap[u.Tid]
		if len(replies) == 0 {
			continue
		}
		firstReply := replies[0]
		lastReply := replies[len(replies)-1]
		jolReplyMap[u.Tid] = replies[1:]
		userId, err := GetMigrateUserService().getUserIdByUsername(ctx, u.AuthorId)
		if err != nil {
			return nil, nil, metaerror.Wrap(err, "get userId failed, userId: %s", u.AuthorId)
		}

		finalDiscuss := foundationmodel.NewDiscussBuilder().
			Id(u.Tid).
			Title(u.Title).
			Content(firstReply.Content).
			AuthorId(userId).
			InsertTime(firstReply.Time).
			ModifyTime(firstReply.Time).
			UpdateTime(lastReply.Time).
			Build()
		if u.Cid > 0 {
			newContestId := GetMigrateContestService().GetNewContestIdByJol(u.Cid)
			finalDiscuss.ContestId = newContestId
		}
		if u.Pid > 0 {
			newProblemId := GetMigrateProblemService().GetNewProblemId(u.Pid)
			finalDiscuss.ProblemId = &newProblemId
		}
		docs = append(docs, finalDiscuss)
	}

	return docs, jolReplyMap, nil
}

func (s *MigrateDiscussService) processCodeojBlog(ctx context.Context) ([]*foundationmodel.Discuss, error) {
	slog.Info("migrate Discuss processCodeojDiscuss")

	db := metamysql.GetSubsystem().GetClient("codeoj")

	var codeojBlogs []CodeojBlog
	if err := db.Where("blog_id < 13").Order("blog_id ASC").Find(&codeojBlogs).Error; err != nil {
		return nil, metaerror.Wrap(err, "query codeojBlogs failed")
	}

	var docs []*foundationmodel.Discuss
	for _, u := range codeojBlogs {
		userId, err := GetMigrateUserService().getUserIdByUsername(ctx, u.Creator)
		if err != nil {
			return nil, metaerror.Wrap(err, "get userId failed, userId: %s", u.Creator)
		}

		discuss := foundationmodel.NewDiscussBuilder().
			Title(u.Title).
			Content(u.Content).
			AuthorId(userId).
			InsertTime(u.InsertTime).
			ModifyTime(u.UpdateTime).
			UpdateTime(u.UpdateTime).
			Build()

		docs = append(docs, discuss)
	}

	return docs, nil
}
