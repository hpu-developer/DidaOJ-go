package foundationservice

import (
	"context"
	"foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"github.com/gin-gonic/gin"
	"meta/singleton"
	"time"
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
	if judgeJob == nil {
		return nil, nil
	}
	user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, judgeJob.AuthorId)
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

func (s *JudgeService) GetJudgeCode(ctx context.Context, id int) (foundationjudge.JudgeLanguage, *string, error) {
	return foundationdao.GetJudgeJobDao().GetJudgeCode(ctx, id)
}

func (s *JudgeService) GetJudgeList(ctx context.Context, problemId string, username string, language foundationjudge.JudgeLanguage, status foundationjudge.JudgeStatus, page int, pageSize int) ([]*foundationmodel.JudgeJob, int, error) {
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
	judgeJobs, totalCount, err := foundationdao.GetJudgeJobDao().GetJudgeJobList(ctx, problemId, userId, language, status, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(judgeJobs) > 0 {
		var userIds []int
		for _, judgeJob := range judgeJobs {
			userIds = append(userIds, judgeJob.AuthorId)
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
			if user, ok := userMap[judgeJob.AuthorId]; ok {
				judgeJob.AuthorUsername = &user.Username
				judgeJob.AuthorNickname = &user.Nickname
			}
		}
	}
	return judgeJobs, totalCount, nil
}

func (s *JudgeService) GetRankAcProblem(ctx *gin.Context, approveStartTime *time.Time, approveEndTime *time.Time, page int, pageSize int) ([]*foundationmodel.UserRank, int, error) {
	rankUsers, totalCount, err := foundationdao.GetJudgeJobDao().GetRankAcProblem(ctx, approveStartTime, approveEndTime, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	if len(rankUsers) > 0 {
		var userIds []int
		for _, rankUser := range rankUsers {
			userIds = append(userIds, rankUser.Id)
		}
		users, err := foundationdao.GetUserDao().GetUsersRankInfo(ctx, userIds)
		if err != nil {
			return nil, 0, err
		}
		userMap := make(map[int]*foundationmodel.UserRankInfo)
		for _, user := range users {
			userMap[user.Id] = user
		}
		for _, rankUser := range rankUsers {
			if user, ok := userMap[rankUser.Id]; ok {
				rankUser.Username = user.Username
				rankUser.Nickname = user.Nickname
				rankUser.Slogan = user.Slogan
			}
		}
	}
	return rankUsers, totalCount, nil
}

func (s *JudgeService) GetUserAcProblemIds(ctx context.Context, userId int) ([]string, error) {
	problemIds, err := foundationdao.GetJudgeJobDao().GetUserAcProblemIds(ctx, userId)
	if err != nil {
		return nil, err
	}
	return problemIds, nil
}

func (s *JudgeService) UpdateJudge(ctx context.Context, id int, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdao.GetJudgeJobDao().UpdateJudgeJob(ctx, id, judgeJob)
}

func (s *JudgeService) InsertJudgeJob(ctx context.Context, judgeJob *foundationmodel.JudgeJob) error {
	return foundationdao.GetJudgeJobDao().InsertJudgeJob(ctx, judgeJob)
}

func (s *JudgeService) RejudgeJob(ctx context.Context, id int) error {
	return foundationdao.GetJudgeJobDao().RejudgeJob(ctx, id)
}

func (s *JudgeService) PostRejudgeProblem(ctx context.Context, id string) error {
	return foundationdao.GetJudgeJobDao().RejudgeProblem(ctx, id)
}

func (s *JudgeService) RejudgeRecently(ctx context.Context) error {
	return foundationdao.GetJudgeJobDao().RejudgeRecently(ctx)
}

func (s *JudgeService) RejudgeAll(ctx context.Context) error {
	return foundationdao.GetJudgeJobDao().RejudgeAll(ctx)
}
