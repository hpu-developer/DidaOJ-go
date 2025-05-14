package foundationservice

import (
	"context"
	"foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"meta/singleton"
	"web/request"
)

type ProblemService struct {
}

var singletonProblemService = singleton.Singleton[ProblemService]{}

func GetProblemService() *ProblemService {
	return singletonProblemService.GetInstance(
		func() *ProblemService {
			return &ProblemService{}
		},
	)
}

func (s *ProblemService) GetProblem(ctx context.Context, id string) (*foundationmodel.Problem, error) {
	problem, err := foundationdao.GetProblemDao().GetProblem(ctx, id)
	if err != nil {
		return nil, err
	}
	if problem == nil {
		return nil, nil
	}
	if problem.CreatorId > 0 {
		user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, problem.CreatorId)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, nil
		}
		problem.CreatorUsername = &user.Username
		problem.CreatorNickname = &user.Nickname
	}
	return problem, nil
}

func (s *ProblemService) GetProblemJudge(ctx context.Context, id string) (*foundationmodel.Problem, error) {
	problem, err := foundationdao.GetProblemDao().GetProblemJudge(ctx, id)
	if err != nil {
		return nil, err
	}
	if problem == nil {
		return nil, nil
	}
	if problem.CreatorId > 0 {
		user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, problem.CreatorId)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, nil
		}
		problem.CreatorUsername = &user.Username
		problem.CreatorNickname = &user.Nickname
	}
	return problem, nil
}

func (s *ProblemService) HasProblem(ctx context.Context, id string) (bool, error) {
	return foundationdao.GetProblemDao().HasProblem(ctx, id)
}

func (s *ProblemService) GetProblemList(ctx context.Context, title string, tag string, page int, pageSize int) ([]*foundationmodel.Problem, int, error) {
	var tags []int
	if tag != "" {
		var err error
		tags, err = foundationdao.GetProblemTagDao().SearchTags(ctx, tag)
		if err != nil {
			return nil, 0, err
		}
		if len(tags) == 0 {
			return nil, 0, nil
		}
	}
	return foundationdao.GetProblemDao().GetProblemList(ctx, title, tags, page, pageSize)
}

func (s *ProblemService) GetProblemListWithUser(ctx context.Context, userId int, title string, tag string, page int, pageSize int,
) ([]*foundationmodel.Problem, int, map[string]foundationmodel.ProblemAttemptStatus, error) {
	var tags []int
	if tag != "" {
		var err error
		tags, err = foundationdao.GetProblemTagDao().SearchTags(ctx, tag)
		if err != nil {
			return nil, 0, nil, err
		}
	}
	problemList, totalCount, err := foundationdao.GetProblemDao().GetProblemList(ctx, title, tags, page, pageSize)
	if err != nil {
		return nil, 0, nil, err
	}
	var problemIds []string
	for _, problem := range problemList {
		problemIds = append(problemIds, problem.Id)
	}
	problemStatus, err := foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(ctx, problemIds, userId)
	if err != nil {
		return nil, 0, nil, err
	}
	return problemList, totalCount, problemStatus, nil
}

func (s *ProblemService) GetProblemTagList(ctx context.Context, maxCount int) ([]*foundationmodel.ProblemTag, int, error) {
	return foundationdao.GetProblemTagDao().GetProblemTagList(ctx, maxCount)
}

func (s *ProblemService) GetProblemTagByIds(ctx context.Context, ids []int) ([]*foundationmodel.ProblemTag, error) {
	return foundationdao.GetProblemTagDao().GetProblemTagByIds(ctx, ids)
}

func (s *ProblemService) PostEdit(ctx context.Context, userId int, requestData *request.ProblemEdit) error {
	return foundationdao.GetProblemDao().PostEdit(ctx, userId, requestData)
}

func (s *ProblemService) UpdateJudgeMd5(ctx context.Context, id string, md5 string) error {
	return foundationdao.GetProblemDao().UpdateJudgeMd5(ctx, id, md5)
}
