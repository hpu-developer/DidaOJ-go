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
	return discuss, err
}

func (s *DiscussService) GetDiscussList(ctx context.Context, page int, pageSize int) ([]*foundationmodel.Discuss, int, error) {
	discusses, totalCount, err := foundationdao.GetDiscussDao().GetDiscussList(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
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
