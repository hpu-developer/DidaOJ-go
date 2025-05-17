package foundationservice

import (
	"context"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	"meta/singleton"
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

func (s *DiscussService) GetDiscuss(ctx context.Context, id int) (*foundationmodel.Discuss, error) {
	discuss, err := foundationdao.GetDiscussDao().GetDiscuss(ctx, id)
	if err != nil {
		return nil, err
	}
	if discuss == nil {
		return nil, nil
	}
	if discuss.AuthorId > 0 {
		user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, discuss.AuthorId)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, nil
		}
		discuss.AuthorUsername = &user.Username
		discuss.AuthorNickname = &user.Nickname
	}
	if discuss.ContestId > 0 {
		// 校验权限

		contestTitle, err := foundationdao.GetContestDao().GetContestTitle(ctx, discuss.ContestId)
		if err != nil {
			return nil, err
		}
		discuss.ContestTitle = contestTitle
		if discuss.ProblemId != nil {
			problemTitle, err := foundationdao.GetProblemDao().GetProblemTitle(ctx, discuss.ProblemId)
			if err != nil {
				return nil, err
			}
			discuss.ProblemTitle = problemTitle
			discuss.ContestProblemIndex, err = foundationdao.GetContestDao().GetProblemIndex(ctx, discuss.ContestId, discuss.ProblemId)
			if err != nil {
				return nil, err
			}
			// 隐藏真实的ProblemId
			discuss.ProblemId = nil
		}
	} else if discuss.ProblemId != nil {
		// 校验权限

		title, err := foundationdao.GetProblemDao().GetProblemTitle(ctx, discuss.ProblemId)
		if err != nil {
			return nil, err
		}
		discuss.ProblemTitle = title
	} else if discuss.JudgeId > 0 {
		// 校验权限

	}

	return discuss, err
}

func (s *DiscussService) GetDiscussList(ctx context.Context, contestId int,
	contestProblemIndex int, problemId string, title string, username string,
	page int, pageSize int) ([]*foundationmodel.Discuss, int, error) {
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
	if contestId > 0 {
		// 计算ProblemId
		if contestProblemIndex > 0 {
			problemIdPtr, err := foundationdao.GetContestDao().GetProblemIdByContest(ctx, contestId, contestProblemIndex)
			if err != nil {
				return nil, 0, err
			}
			if problemIdPtr == nil {
				return nil, 0, nil
			}
			problemId = *problemIdPtr
		}
	}

	discusses, totalCount, err := foundationdao.GetDiscussDao().GetDiscussList(ctx, contestId,
		problemId, title, userId,
		page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(discusses) > 0 {
		var userIds []int
		for _, discuss := range discusses {
			userIds = append(userIds, discuss.AuthorId)
		}
		users, err := foundationdao.GetUserDao().GetUsersAccountInfo(ctx, userIds)
		if err != nil {
			return nil, 0, err
		}
		userMap := make(map[int]*foundationmodel.UserAccountInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, discuss := range discusses {
			if user, ok := userMap[discuss.AuthorId]; ok {
				discuss.AuthorUsername = &user.Username
				discuss.AuthorNickname = &user.Nickname
			}
		}

		for _, discuss := range discusses {
			if discuss.ContestId > 0 {
				if discuss.ProblemId != nil {
					problemTitle, err := foundationdao.GetProblemDao().GetProblemTitle(ctx, discuss.ProblemId)
					if err != nil {
						return nil, 0, err
					}
					discuss.ProblemTitle = problemTitle
					discuss.ContestProblemIndex, err = foundationdao.GetContestDao().GetProblemIndex(ctx, discuss.ContestId, discuss.ProblemId)
					if err != nil {
						return nil, 0, err
					}
					// 隐藏真实的ProblemId
					discuss.ProblemId = nil
				}
			}
		}
	}
	return discusses, totalCount, nil
}

func (s *DiscussService) GetDiscussTagByIds(ctx *gin.Context, tags []int) ([]*foundationmodel.DiscussTag, error) {
	return foundationdao.GetDiscussTagDao().GetDiscussTagByIds(ctx, tags)
}

func (s *DiscussService) InsertDiscuss(ctx context.Context, discuss *foundationmodel.Discuss) error {
	return foundationdao.GetDiscussDao().InsertDiscuss(ctx, discuss)
}

func (s *DiscussService) GetDiscussCommentList(ctx *gin.Context, discussComment int, page int, pageSize int) ([]*foundationmodel.DiscussComment, int, error) {
	discussComments, totalCount, err := foundationdao.GetDiscussCommentDao().GetDiscussCommentList(ctx, discussComment, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(discussComments) > 0 {
		var userIds []int
		for _, discussComment := range discussComments {
			userIds = append(userIds, discussComment.AuthorId)
		}
		users, err := foundationdao.GetUserDao().GetUsersAccountInfo(ctx, userIds)
		if err != nil {
			return nil, 0, err
		}
		userMap := make(map[int]*foundationmodel.UserAccountInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, discussComment := range discussComments {
			if user, ok := userMap[discussComment.AuthorId]; ok {
				discussComment.AuthorUsername = &user.Username
				discussComment.AuthorNickname = &user.Nickname
			}
		}
	}
	return discussComments, totalCount, nil
}
