package service

import (
	"context"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"log/slog"
	metaerror "meta/meta-error"
	metamysql "meta/meta-mysql"
	"meta/singleton"
	migratedao "migrate/dao"
	"slices"
	"sort"
	"strconv"
	"time"
)

type MigrateDiscussEojService struct{}

var singletonMigrateDiscussEojService = singleton.Singleton[MigrateDiscussEojService]{}

func GetMigrateDiscussEojService() *MigrateDiscussEojService {
	return singletonMigrateDiscussEojService.GetInstance(
		func() *MigrateDiscussEojService {
			return &MigrateDiscussEojService{}
		},
	)
}

type EojBlog struct {
	Id         int       `gorm:"column:id"`
	Title      string    `gorm:"column:title"`
	Text       string    `gorm:"column:text"`
	CreateTime time.Time `gorm:"column:create_time"`
	EditTime   time.Time `gorm:"column:edit_time"`
	AuthorId   int       `gorm:"column:author_id"`
}

func (EojBlog) TableName() string {
	return "blog_blog"
}

type EojComment struct {
	Id            int       `gorm:"column:id"`
	ObjectPk      int       `gorm:"column:object_pk"`
	Comment       string    `gorm:"column:comment"`
	SubmitDate    time.Time `gorm:"column:submit_date"`
	IsRemoved     int       `gorm:"column:is_removed"`
	ContentTypeId int       `gorm:"column:content_type_id"`
	UserId        int       `gorm:"column:user_id"`
}

func (EojComment) TableName() string {
	return "django_comments"
}

type EojContestClarification struct {
	Id        int       `gorm:"column:id"`
	Text      string    `gorm:"column:text"`
	Time      time.Time `gorm:"column:time"`
	Important int       `gorm:"column:important"`
	Answer    string    `gorm:"column:answer"`
	AuthorId  int       `gorm:"column:author_id"`
	ContestId int       `gorm:"column:contest_id"`
}

func (EojContestClarification) TableName() string {
	return "contest_contestclarification"
}

func (s *MigrateDiscussEojService) Start() error {
	ctx := context.Background()

	var discusss []*foundationmodel.Discuss

	eojDiscusss, err := s.processDiscussEojBlog(ctx)
	if err != nil {
		return err
	}
	discusss = append(discusss, eojDiscusss...)

	eojProblemComments, err := s.processDiscussEojProblemComments(ctx)
	if err != nil {
		return err
	}
	discusss = append(discusss, eojProblemComments...)

	eojContestDiscusss, err := s.processDiscussEojContests(ctx)
	if err != nil {
		return err
	}
	discusss = append(discusss, eojContestDiscusss...)

	didaojDiscusss, err := s.processDiscussDidaOjDiscuss(ctx)
	if err != nil {
		return err
	}
	discusss = append(discusss, didaojDiscusss...)

	slog.Info("migrate discuss updates", "count", len(discusss))

	sort.Slice(
		discusss, func(i, j int) bool {
			return discusss[i].InsertTime.Before(discusss[j].InsertTime)
		},
	)

	for _, discuss := range discusss {
		err = foundationdao.GetDiscussDao().InsertDiscuss(ctx, discuss)
		if err != nil {
			return metaerror.Wrap(err, "insert discuss failed")
		}
	}

	didaojIdMap := map[int]int{}
	eojBlogIdMap := map[int]int{}
	eojClarificationIdMap := map[int]int{}
	for _, discuss := range discusss {
		if discuss.MigrateDidaOJId > 0 {
			didaojIdMap[discuss.MigrateDidaOJId] = discuss.Id
		} else if discuss.MigrateEojBlogId > 0 {
			eojBlogIdMap[discuss.MigrateEojBlogId] = discuss.Id
		} else if discuss.MigrateEojClarificationId > 0 {
			eojClarificationIdMap[discuss.MigrateEojClarificationId] = discuss.Id
		}
	}

	var discussComments []*foundationmodel.DiscussComment

	comments1, err := s.processEojBlogComments(ctx)
	if err != nil {
		return err
	}
	discussComments = append(discussComments, comments1...)

	comments2, err := s.processEojContestComments(ctx)
	if err != nil {
		return err
	}
	discussComments = append(discussComments, comments2...)

	didaojDiscussComments, err := s.processDidaOjDiscussComment(ctx)
	if err != nil {
		return err
	}
	discussComments = append(discussComments, didaojDiscussComments...)

	sort.Slice(
		discussComments, func(i, j int) bool {
			return discussComments[i].InsertTime.Before(discussComments[j].InsertTime)
		},
	)

	for _, discussComment := range discussComments {
		if discussComment.MigrateDidaOJId > 0 {
			discussComment.DiscussId = didaojIdMap[discussComment.MigrateDidaOJId]
		} else if discussComment.MigrateEojBlogId > 0 {
			discussComment.DiscussId = eojBlogIdMap[discussComment.MigrateEojBlogId]
		} else if discussComment.MigrateEojClarificationId > 0 {
			discussComment.DiscussId = eojClarificationIdMap[discussComment.MigrateEojClarificationId]
		}
	}

	for _, discussComment := range discussComments {
		if discussComment.DiscussId <= 0 {
			return metaerror.New("discuss comment discuss id is not set, maybe the discuss is not migrated")
		}
		err = foundationdao.GetDiscussCommentDao().InsertDiscussComment(ctx, discussComment)
		if err != nil {
			return metaerror.Wrap(err, "insert discuss failed")
		}
	}

	return nil
}

func (s *MigrateDiscussEojService) processDiscussEojBlog(ctx context.Context) ([]*foundationmodel.Discuss, error) {

	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	ignoreIds := []int{39}

	const batchSize = 1000
	offset := 0

	slog.Info("migrate discuss start", "batchSize", batchSize)

	var discusss []*foundationmodel.Discuss

	for {
		var discussModels []EojBlog
		if err := eojDb.
			Select("id, title, text, create_time, edit_time, author_id").
			Model(&EojBlog{}).
			Order("id ASC").
			Limit(batchSize).
			Offset(offset).
			Find(&discussModels).Error; err != nil {
			return nil, metaerror.Wrap(err, "query eoj problem managers failed")
		}
		if len(discussModels) == 0 {
			break
		}

		for _, discussModel := range discussModels {
			if slices.Contains(ignoreIds, discussModel.Id) {
				continue
			}
			userId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(discussModel.AuthorId))
			if err != nil {
				return nil, metaerror.Wrap(err, "get eoj user mark failed")
			}
			realUserId, err := strconv.Atoi(*userId)
			if err != nil {
				return nil, metaerror.Wrap(err, "convert eoj user id to int failed")
			}
			discuss := foundationmodel.NewDiscussBuilder().
				Title(discussModel.Title).
				Content(discussModel.Text).
				AuthorId(realUserId).
				InsertTime(discussModel.CreateTime).
				ModifyTime(discussModel.EditTime).
				UpdateTime(discussModel.EditTime).
				Build()

			discuss.MigrateEojBlogId = discussModel.Id

			discusss = append(discusss, discuss)
		}

		slog.Info("migrate discuss", "offset", offset, "batchSize", batchSize)

		offset += batchSize
	}

	return discusss, nil
}

func (s *MigrateDiscussEojService) processDiscussEojProblemComments(ctx context.Context) (
	[]*foundationmodel.Discuss,
	error,
) {

	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	var discusss []*foundationmodel.Discuss

	// 查询所有唯一标签
	var comments []EojComment
	if err := eojDb.
		Model(&EojComment{}).
		Where("content_type_id = 12").
		Where("is_removed = 0").
		Scan(&comments).Error; err != nil {
		return nil, metaerror.Wrap(err, "query problem_tag failed")
	}
	for _, discussModel := range comments {
		userId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(discussModel.UserId))
		if err != nil {
			return nil, metaerror.Wrap(err, "get eoj user mark failed")
		}
		realUserId, err := strconv.Atoi(*userId)
		if err != nil {
			return nil, metaerror.Wrap(err, "convert eoj user id to int failed")
		}

		title := discussModel.Comment
		content := discussModel.Comment

		oldProblemId := discussModel.ObjectPk
		problemId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-problem", strconv.Itoa(oldProblemId))
		if err != nil {
			return nil, metaerror.Wrap(err, "get eoj problem mark failed")
		}

		if len(content) > 20 {
			title = fmt.Sprintf("Problem [%s] Comment", *problemId)
		}

		discuss := foundationmodel.NewDiscussBuilder().
			Title(title).
			Content(content).
			AuthorId(realUserId).
			InsertTime(discussModel.SubmitDate).
			ModifyTime(discussModel.SubmitDate).
			UpdateTime(discussModel.SubmitDate).
			ProblemId(problemId).
			Build()

		discusss = append(discusss, discuss)
	}

	return discusss, nil
}

func (s *MigrateDiscussEojService) processDiscussEojContests(ctx context.Context) (
	[]*foundationmodel.Discuss,
	error,
) {

	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	var discusss []*foundationmodel.Discuss

	// 查询所有唯一标签
	var clarifications []EojContestClarification
	if err := eojDb.
		Model(&EojContestClarification{}).
		Scan(&clarifications).Error; err != nil {
		return nil, metaerror.Wrap(err, "query problem_tag failed")
	}
	for _, discussModel := range clarifications {
		userId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(discussModel.AuthorId))
		if err != nil {
			return nil, metaerror.Wrap(err, "get eoj user mark failed")
		}
		realUserId, err := strconv.Atoi(*userId)
		if err != nil {
			return nil, metaerror.Wrap(err, "convert eoj user id to int failed")
		}

		var title string
		var content string

		if discussModel.Text == "" {
			title = discussModel.Answer
			if len(title) > 20 {
				title = fmt.Sprintf("%s...", title[:20])
			}
			content = discussModel.Answer
		} else {
			title = discussModel.Text
			if len(title) > 20 {
				title = fmt.Sprintf("%s...", title[:20])
			}
			content = discussModel.Text
		}

		oldContestId := discussModel.ContestId
		contestIdStr, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-contest", strconv.Itoa(oldContestId))
		if err != nil {
			return nil, metaerror.Wrap(err, "get eoj problem mark failed")
		}
		contestId, err := strconv.Atoi(*contestIdStr)
		if err != nil {
			return nil, metaerror.Wrap(err, "convert eoj contest id to int failed")
		}

		discuss := foundationmodel.NewDiscussBuilder().
			Title(title).
			Content(content).
			AuthorId(realUserId).
			InsertTime(discussModel.Time).
			ModifyTime(discussModel.Time).
			UpdateTime(discussModel.Time).
			ContestId(contestId).
			Build()

		discuss.MigrateEojClarificationId = discussModel.Id

		discusss = append(discusss, discuss)
	}

	return discusss, nil
}

func (s *MigrateDiscussEojService) processDiscussDidaOjDiscuss(ctx context.Context) (
	[]*foundationmodel.Discuss,
	error,
) {
	list, err := migratedao.GetDiscussDao().GetDiscussList(ctx)
	for _, discuss := range list {
		discuss.MigrateDidaOJId = discuss.Id
	}
	return list, err
}

func (s *MigrateDiscussEojService) processEojBlogComments(ctx context.Context) (
	[]*foundationmodel.DiscussComment,
	error,
) {

	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	var discussComments []*foundationmodel.DiscussComment

	// 查询所有唯一标签
	var comments []EojComment
	if err := eojDb.
		Model(&EojComment{}).
		Where("content_type_id = 34").
		Where("is_removed = 0").
		Scan(&comments).Error; err != nil {
		return nil, metaerror.Wrap(err, "query problem_tag failed")
	}

	for _, commentModel := range comments {
		userId, err := migratedao.GetMigrateMarkDao().GetMark(ctx, "eoj-user", strconv.Itoa(commentModel.UserId))
		if err != nil {
			return nil, metaerror.Wrap(err, "get eoj user mark failed")
		}
		realUserId, err := strconv.Atoi(*userId)
		if err != nil {
			return nil, metaerror.Wrap(err, "convert eoj user id to int failed")
		}

		discussComment := foundationmodel.NewDiscussCommentBuilder().
			DiscussId(commentModel.ObjectPk).
			Content(commentModel.Comment).
			AuthorId(realUserId).
			InsertTime(commentModel.SubmitDate).
			UpdateTime(commentModel.SubmitDate).
			Build()

		discussComment.MigrateEojBlogId = commentModel.Id

		discussComments = append(discussComments, discussComment)
	}

	return discussComments, nil
}

func (s *MigrateDiscussEojService) processEojContestComments(ctx context.Context) (
	[]*foundationmodel.DiscussComment,
	error,
) {

	eojDb := metamysql.GetSubsystem().GetClient("eoj")

	var discussComments []*foundationmodel.DiscussComment

	// 查询所有唯一标签
	var clarifications []EojContestClarification
	if err := eojDb.
		Model(&EojContestClarification{}).
		Where("text != ''").
		Where("answer != ''").
		Scan(&clarifications).Error; err != nil {
		return nil, metaerror.Wrap(err, "query problem_tag failed")
	}

	for _, commentModel := range clarifications {
		realContestIdStr, err := migratedao.GetMigrateMarkDao().GetMark(
			ctx,
			"eoj-contest",
			strconv.Itoa(commentModel.ContestId),
		)
		if err != nil {
			return nil, metaerror.Wrap(err, "get eoj user mark failed")
		}
		realContestId, err := strconv.Atoi(*realContestIdStr)
		if err != nil {
			return nil, metaerror.Wrap(err, "convert eoj user id to int failed")
		}
		ownerId, err := foundationdao.GetContestDao().GetContestOwnerId(ctx, realContestId)
		if err != nil {
			return nil, err
		}

		discussComment := foundationmodel.NewDiscussCommentBuilder().
			Content(commentModel.Answer).
			AuthorId(ownerId).
			InsertTime(commentModel.Time).
			UpdateTime(commentModel.Time).
			Build()

		discussComment.MigrateEojClarificationId = commentModel.Id

		discussComments = append(discussComments, discussComment)
	}

	return discussComments, nil
}

func (s *MigrateDiscussEojService) processDidaOjDiscussComment(ctx context.Context) (
	[]*foundationmodel.DiscussComment,
	error,
) {
	list, err := migratedao.GetDiscussCommentDao().GetDiscussCommentList(ctx)
	for _, discuss := range list {
		discuss.MigrateDidaOJId = discuss.Id
	}
	return list, err
}
