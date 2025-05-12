package foundationservice

import (
	"context"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"meta/singleton"
)

type JudgeService struct {
}

var singletonJudgeService = singleton.Singleton[JudgeService]{}

func GetJudgeService() *JudgeService {
	return singletonJudgeService.GetInstance(
		func() *JudgeService {
			return &JudgeService{}
		},
	)
}

func (s *JudgeService) GetJudge(ctx context.Context, id int) (*foundationmodel.JudgeJob, error) {
	judgeJob, err := foundationdao.GetJudgeJobDao().GetJudgeJob(ctx, id)
	if err != nil {
		return nil, err
	}
	user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, judgeJob.Author)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, nil
	}
	judgeJob.AuthorUsername = &user.Username
	judgeJob.AuthorNickname = &user.Nickname
	return judgeJob, nil
}

func (s *JudgeService) GetJudgeList(ctx context.Context, page int, pageSize int) ([]*foundationmodel.JudgeJob, int, error) {
	judgeJobs, totalCount, err := foundationdao.GetJudgeJobDao().GetJudgeJobList(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(judgeJobs) > 0 {
		var userIds []int
		for _, judgeJob := range judgeJobs {
			userIds = append(userIds, judgeJob.Author)
		}
		users, err := foundationdao.GetUserDao().GetUsersAccountInfo(ctx, userIds)
		if err != nil {
			return nil, 0, err
		}
		userMap := make(map[int]*foundationmodel.UserAccountInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, judgeJob := range judgeJobs {
			if user, ok := userMap[judgeJob.Author]; ok {
				judgeJob.AuthorUsername = &user.Username
				judgeJob.AuthorNickname = &user.Nickname
			}
		}
	}
	return judgeJobs, totalCount, nil
}

func (s *JudgeService) UpdateJudge(ctx context.Context, id int, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdao.GetJudgeJobDao().UpdateJudgeJob(ctx, id, judgeJob)
}

func (s *JudgeService) InsertJudgeJob(ctx context.Context, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdao.GetJudgeJobDao().InsertJudgeJob(ctx, judgeJob)
}

func (s *JudgeService) RejudgeRecently(ctx context.Context) error {
	return foundationdao.GetJudgeJobDao().RejudgeRecently(ctx)
}
