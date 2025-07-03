package foundationservice

import (
	"context"
	foundationauth "foundation/foundation-auth"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"github.com/gin-gonic/gin"
	"meta/singleton"
	"time"
)

type DiscussService struct {
}

var singletonDiscussService = singleton.Singleton[DiscussService]{}

func GetDiscussService() *DiscussService {
	return singletonDiscussService.GetInstance(
		func() *DiscussService {
			return &DiscussService{}
		},
	)
}

func (s *DiscussService) CheckEditAuth(ctx *gin.Context, id int) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageDiscuss)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		ownerId, err := foundationdao.GetDiscussDao().GetInserter(ctx, id)
		if err != nil {
			return userId, false, err
		}
		if ownerId != userId {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *DiscussService) CheckViewAuth(ctx *gin.Context, id int) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageDiscuss)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		isBanned, err := foundationdao.GetDiscussDao().IsDiscussBannedOrNotExist(ctx, id)
		if err != nil {
			return userId, false, err
		}
		if isBanned {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *DiscussService) CheckEditCommentAuth(ctx *gin.Context, commentId int) (
	int,
	bool,
	*foundationview.DiscussCommentViewEdit,
	error,
) {
	discussComment, err := foundationdao.GetDiscussCommentDao().GetCommentEditView(ctx, commentId)
	if err != nil {
		return 0, false, nil, err
	}
	if discussComment == nil {
		return 0, false, nil, nil
	}
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageDiscuss)
	if err != nil {
		return userId, false, discussComment, err
	}
	if userId <= 0 {
		return userId, false, discussComment, nil
	}
	if !hasAuth {
		isBanned, err := foundationdao.GetDiscussDao().IsDiscussBannedOrNotExist(ctx, discussComment.DiscussId)
		if err != nil {
			return userId, false, discussComment, err
		}
		if isBanned {
			return userId, false, discussComment, nil
		}
		if discussComment.Inserter != userId {
			return userId, false, discussComment, nil
		}
	}
	return userId, true, discussComment, nil
}

func (s *DiscussService) GetContent(ctx context.Context, id int) (*string, error) {
	return foundationdao.GetDiscussDao().GetContent(ctx, id)
}

func (s *DiscussService) GetDiscussView(ctx context.Context, id int) (*foundationview.DiscussDetail, error) {
	discuss, err := foundationdao.GetDiscussDao().GetDiscussDetail(ctx, id)
	if err != nil {
		return nil, err
	}
	if discuss == nil {
		return nil, nil
	}
	if discuss.ContestId != nil {
		// 校验权限
		discuss.ContestTitle, err = foundationdao.GetContestDao().GetContestTitle(ctx, *discuss.ContestId)
		if err != nil {
			return nil, err
		}
		if discuss.ProblemId != nil {
			problemTitle, err := foundationdao.GetProblemDao().GetProblemTitle(ctx, *discuss.ProblemId)
			if err != nil {
				return nil, err
			}
			discuss.ProblemTitle = problemTitle
			discuss.ContestProblemIndex, err = foundationdao.GetContestProblemDao().GetProblemIndex(
				ctx,
				*discuss.ContestId,
				*discuss.ProblemId,
			)
			if err != nil {
				return nil, err
			}
			// 隐藏真实的ProblemId
			discuss.ProblemId = nil
		}
	} else if discuss.ProblemId != nil {
		// 校验权限
		title, err := foundationdao.GetProblemDao().GetProblemTitle(ctx, *discuss.ProblemId)
		if err != nil {
			return nil, err
		}
		discuss.ProblemTitle = title
	} else if discuss.JudgeId != nil {
		// 校验权限

	}

	return discuss, err
}

func (s *DiscussService) GetDiscussList(
	ctx context.Context,
	onlyProblem bool,
	contestId int,
	problemId int,
	title string,
	username string,
	page int,
	pageSize int,
) ([]*foundationview.DiscussList, int, error) {
	var err error
	userId := -1
	if username != "" {
		userId, err = foundationdao.GetUserDao().GetUserIdByUsername(ctx, username)
		if err != nil {
			return nil, 0, err
		}
		if userId <= 0 {
			return nil, 0, nil
		}
	}

	discusses, totalCount, err := foundationdao.GetDiscussDao().GetDiscussList(
		ctx, onlyProblem, contestId, problemId, title, userId,
		page, pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	if contestId > 0 {
		for _, discuss := range discusses {
			discuss.ProblemId = nil
		}
	}
	return discusses, totalCount, nil
}

func (s *DiscussService) GetDiscussTags(ctx *gin.Context, id int) ([]*foundationmodel.DiscussTag, error) {
	return foundationdao.GetDiscussTagDao().GetDiscussTags(ctx, id)
}

func (s *DiscussService) InsertDiscuss(ctx context.Context, discuss *foundationmodel.Discuss) error {
	return foundationdao.GetDiscussDao().InsertDiscuss(ctx, discuss)
}

func (s *DiscussService) InsertDiscussComment(ctx context.Context, discuss *foundationmodel.DiscussComment) error {
	return foundationdao.GetDiscussCommentDao().InsertDiscussComment(ctx, discuss)
}

func (s *DiscussService) GetDiscussCommentList(
	ctx *gin.Context,
	discussComment int,
	page int,
	pageSize int,
) ([]*foundationview.DiscussCommentList, int, error) {
	discussComments, totalCount, err := foundationdao.GetDiscussCommentDao().GetDiscussCommentList(
		ctx,
		discussComment,
		page,
		pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	return discussComments, totalCount, nil
}

func (s *DiscussService) PostEdit(ctx *gin.Context, id int, discuss *foundationmodel.Discuss) error {
	return foundationdao.GetDiscussDao().PostEdit(ctx, id, discuss)
}

func (s *DiscussService) UpdateContent(ctx context.Context, id int, content string) error {
	return foundationdao.GetDiscussDao().UpdateContent(ctx, id, content)
}

func (s *DiscussService) UpdateCommentContentAndTime(
	ctx context.Context,
	userId int,
	commentId int, discussId int, content string,
	updateTime time.Time,
) error {
	return foundationdao.GetDiscussCommentDao().UpdateContent(ctx, userId, commentId, discussId, content, updateTime)
}
