package foundationservice

import (
	"bytes"
	"context"
	"encoding/json"
	foundationauth "foundation/foundation-auth"
	foundationcontest "foundation/foundation-contest"
	foundationdao "foundation/foundation-dao"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"io"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
	metapath "meta/meta-path"
	metazip "meta/meta-zip"
	"meta/singleton"
	"mime/multipart"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ContestService struct {
}

var singletonContestService = singleton.Singleton[ContestService]{}

func GetContestService() *ContestService {
	return singletonContestService.GetInstance(
		func() *ContestService {
			return &ContestService{}
		},
	)
}

func (s *ContestService) CheckEditAuth(ctx *gin.Context, id int) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageContest)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		ownerId, err := foundationdao.GetContestDao().GetContestInserter(ctx, id)
		if err != nil {
			return userId, false, err
		}
		if ownerId != userId {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

// CheckViewAuthWithoutStartTime 检查查看比赛权限，不考虑比赛开始时间
func (s *ContestService) CheckViewAuthWithoutStartTime(ctx *gin.Context, id int) (int, bool, error) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageContest)
	if err != nil {
		return userId, false, err
	}
	if !hasAuth {
		hasAuth, err = foundationdao.GetContestDao().HasContestViewAuthWithoutStartTime(ctx, id, userId)
		if err != nil {
			return userId, false, err
		}
		return userId, hasAuth, nil
	}
	return userId, true, nil
}

func (s *ContestService) CheckViewAuthWithEnd(ctx *gin.Context, id int) (int, bool, error) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageContest)
	if err != nil {
		return userId, false, err
	}
	if !hasAuth {
		hasAuth, err = foundationdao.GetContestDao().HasContestViewAuthWithEnd(ctx, id, userId)
		if err != nil {
			return userId, false, err
		}
		return userId, hasAuth, nil
	}
	return userId, true, nil
}

// CheckViewAuth 检查查看比赛权限
func (s *ContestService) CheckViewAuth(ctx *gin.Context, id int) (int, bool, error) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageContest)
	if err != nil {
		return userId, false, err
	}
	if !hasAuth {
		hasAuth, err = foundationdao.GetContestDao().HasContestViewAuth(ctx, id, userId)
		if err != nil {
			return userId, false, err
		}
		return userId, hasAuth, nil
	}
	return userId, true, nil
}

func (s *ContestService) CheckSubmitAuth(ctx *gin.Context, id int) (int, bool, error) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageContest)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		hasAuth, err = foundationdao.GetContestDao().HasContestSubmitAuth(ctx, id, userId)
		if err != nil {
			return userId, false, err
		}
		return userId, hasAuth, nil
	}
	return userId, true, nil
}

func (s *ContestService) HasContestTitle(ctx *gin.Context, userId int, title string) (bool, error) {
	return foundationdao.GetContestDao().HasContestTitle(ctx, userId, title)
}

func (s *ContestService) GetContestDescription(ctx *gin.Context, id int) (*string, error) {
	return foundationdao.GetContestDao().GetContestDescription(ctx, id)
}

func (s *ContestService) GetContest(ctx *gin.Context, id int, nowTime time.Time) (
	contest *foundationview.ContestDetail,
	hasAuth bool,
	needPassword bool,
	attemptStatus map[int]foundationenum.ProblemAttemptStatus,
	err error,
) {
	contest, err = foundationdao.GetContestDao().GetContest(ctx, id)
	if err != nil {
		return
	}
	if contest == nil {
		return
	}

	hasAuth = true
	needPassword = contest.Password != nil && *contest.Password != ""
	contest.Password = nil

	var userId int
	userId, hasAuth, err = GetContestService().CheckViewAuthWithoutStartTime(ctx, id)
	if err != nil {
		return
	}
	if !hasAuth {
		return
	}
	// 如果还没开始，则不需要获取问题和提交状态了
	if nowTime.Before(contest.StartTime) {
		return
	}
	contest.Problems, err = foundationdao.GetContestProblemDao().GetProblemsDetail(ctx, id)
	if err != nil {
		return
	}
	if len(contest.Problems) == 0 {
		return
	}
	var problemIds []int
	contestProblemMap := make(map[int]*foundationview.ContestProblemDetail, len(contest.Problems))
	for _, problem := range contest.Problems {
		problemIds = append(problemIds, problem.ProblemId)
		contestProblemMap[problem.ProblemId] = problem
	}

	attemptInfos, err := foundationdao.GetContestDao().GetProblemAttemptInfo(
		ctx,
		id,
		problemIds,
		&contest.StartTime,
		&contest.EndTime,
	)
	for _, attemptInfo := range attemptInfos {
		if problem, ok := contestProblemMap[attemptInfo.Id]; ok {
			problem.Attempt = attemptInfo.Attempt
			problem.Accept = attemptInfo.Accept
		}
	}

	if userId > 0 {
		var attemptStatusMap map[int]foundationenum.ProblemAttemptStatus
		attemptStatusMap, err = foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
			ctx,
			userId,
			problemIds,
			id,
			&contest.StartTime,
			&contest.EndTime,
		)
		if err != nil {
			return
		}
		for _, problem := range contest.Problems {
			if judgeAccept, ok := attemptStatusMap[problem.Id]; ok {
				problem.Status = judgeAccept
			} else {
				problem.Status = foundationenum.ProblemAttemptStatusNone
			}
		}
	}

	for _, problem := range contest.Problems {
		problem.ProblemId = 0
		problem.ViewId = nil
	}

	return
}

func (s *ContestService) GetContestClone(ctx *gin.Context, id int) (*foundationview.ContestDetailClone, bool, error) {
	_, hasAuth, err := GetContestService().CheckViewAuthWithEnd(ctx, id)
	if err != nil {
		return nil, false, err
	}
	if !hasAuth {
		return nil, false, nil
	}
	contest, err := foundationdao.GetContestDao().GetContestClone(ctx, id)
	if err != nil {
		return nil, false, err
	}
	if contest == nil {
		return nil, false, nil
	}
	contest.Problems, err = foundationdao.GetContestProblemDao().GetProblemIds(ctx, id)
	if err != nil {
		return nil, false, err
	}
	contest.Members, err = foundationdao.GetContestMemberDao().GetUserIds(ctx, id)
	if err != nil {
		return nil, false, err
	}
	return contest, true, nil
}

func (s *ContestService) GetContestEdit(ctx context.Context, id int) (*foundationview.ContestDetailEdit, error) {
	contest, err := foundationdao.GetContestDao().GetContestEdit(ctx, id)
	if err != nil {
		return nil, err
	}
	if contest == nil {
		return nil, nil
	}
	contest.Problems, err = foundationdao.GetContestProblemDao().GetProblemIds(ctx, id)
	if err != nil {
		return nil, err
	}
	contest.Members, err = foundationdao.GetContestMemberDao().GetUsersWithName(ctx, id)
	if err != nil {
		return nil, err
	}
	return contest, err
}

func (s *ContestService) GetContestStatistics(ctx *gin.Context, id int, language foundationjudge.JudgeLanguage) (
	[]*foundationview.ContestProblemStatistics,
	error,
) {
	statistics, err := foundationdao.GetJudgeJobDao().GetContestStatistics(
		ctx,
		id,
		language,
	)
	if err != nil {
		return nil, err
	}
	for _, stat := range statistics {
		stat.ProblemId = 0
	}
	return statistics, nil
}

func (s *ContestService) GetContestStartTime(ctx *gin.Context, id int) (*time.Time, error) {
	return foundationdao.GetContestDao().GetContestStartTime(ctx, id)
}

func (s *ContestService) GetContestProblems(ctx *gin.Context, id int) (
	[]int,
	error,
) {
	problems, err := foundationdao.GetContestProblemDao().GetProblems(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(problems) == 0 {
		return nil, nil
	}
	var problemIndexes []int
	for _, problem := range problems {
		problemIndexes = append(problemIndexes, int(problem.Index))
	}
	return problemIndexes, nil
}

func (s *ContestService) GetContestProblemsWithAttemptStatus(ctx *gin.Context, id int, userId int) (
	[]int,
	map[int]foundationenum.ProblemAttemptStatus,
	error,
) {
	problems, err := foundationdao.GetContestProblemDao().GetProblems(ctx, id)
	if err != nil {
		return nil, nil, err
	}
	if len(problems) == 0 {
		return nil, nil, nil
	}
	var attemptStatusesMap map[int]foundationenum.ProblemAttemptStatus
	if userId > 0 {
		var problemIds []int
		for _, problem := range problems {
			problemIds = append(problemIds, problem.ProblemId)
		}
		attemptStatuses, err := foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
			ctx,
			userId,
			problemIds,
			id,
			nil,
			nil,
		)
		if err != nil {
			return nil, nil, err
		}
		problemIdMap := make(map[int]int)
		for _, problem := range problems {
			problemIdMap[problem.ProblemId] = int(problem.Index)
		}
		for problemId, attemptStatus := range attemptStatuses {
			if index, ok := problemIdMap[problemId]; ok {
				if attemptStatusesMap == nil {
					attemptStatusesMap = make(map[int]foundationenum.ProblemAttemptStatus)
				}
				attemptStatusesMap[index] = attemptStatus
			}
		}
	}
	var problemIndexes []int
	for _, problem := range problems {
		problemIndexes = append(problemIndexes, int(problem.Index))
	}
	return problemIndexes, attemptStatusesMap, nil
}

func (s *ContestService) GetContestList(
	ctx context.Context,
	title string,
	username string,
	page int,
	pageSize int,
) ([]*foundationview.ContestList, int, error) {
	userId := -1
	if username != "" {
		var err error
		userId, err = foundationdao.GetUserDao().GetUserIdByUsername(ctx, username)
		if err != nil {
			return nil, 0, err
		}
		if userId <= 0 {
			return nil, 0, nil
		}
	}
	return foundationdao.GetContestDao().GetContestList(ctx, title, userId, page, pageSize)
}

func (s *ContestService) GetProblemKeyByContestIndex(ctx context.Context, id int, problemIndex int) (string, error) {
	return foundationdao.GetContestProblemDao().GetProblemKey(ctx, id, problemIndex)
}

func (s *ContestService) GetProblemIdByContestIndex(ctx context.Context, id int, problemIndex int) (int, error) {
	return foundationdao.GetContestProblemDao().GetProblemId(ctx, id, problemIndex)
}

func (s *ContestService) GetProblemIdByContestIndexKey(ctx context.Context, id int, problemKey string) (int, error) {
	constProblemIndex := foundationcontest.GetContestProblemIndex(problemKey)
	return foundationdao.GetContestProblemDao().GetProblemId(ctx, id, constProblemIndex)
}

func (s *ContestService) GetContestProblemIndexById(ctx context.Context, id int, problemId int) (int, error) {
	return foundationdao.GetContestProblemDao().GetProblemIndex(ctx, id, problemId)
}

func (s *ContestService) GetContestRanks(ctx context.Context, id int, nowTime time.Time) (
	contest *foundationview.ContestRankDetail,
	ranks []*foundationview.ContestRank,
	isLocked bool,
	err error,
) {
	contest, err = foundationdao.GetContestDao().GetContestViewRank(ctx, id)
	if err != nil {
		return
	}
	contest.Members, err = foundationdao.GetContestMemberDao().GetUsersWithName(ctx, id)
	if err != nil {
		return
	}
	var contestProblems []*foundationview.ContestProblemRank
	contestProblems, err = foundationdao.GetContestProblemDao().GetProblemsRank(ctx, id)
	if err != nil {
		return
	}

	var problemIndexes []int
	problemMap := make(map[int]uint8)
	for _, problem := range contestProblems {
		problemMap[problem.ProblemId] = problem.Index
		problemIndexes = append(problemIndexes, int(problem.Index))
	}

	isEnd := nowTime.After(contest.EndTime)
	hasLockDuration := contest.LockRankDuration != nil && *contest.LockRankDuration > 0
	isLocked = hasLockDuration &&
		(contest.AlwaysLock || !isEnd) &&
		nowTime.After(contest.EndTime.Add(-*contest.LockRankDuration))

	var lockTimePtr *time.Time
	if isLocked {
		lockTime := contest.EndTime.Add(-*contest.LockRankDuration)
		lockTimePtr = &lockTime
	} else {
		lockTimePtr = nil
	}

	contestRanks, err := foundationdao.GetJudgeJobDao().GetContestRanks(
		ctx, id,
		contest.StartTime,
		lockTimePtr,
		problemMap,
	)
	if err != nil {
		return
	}

	contest.Problems = problemIndexes

	return contest, contestRanks, isLocked, nil
}

func (s *ContestService) GetContestMember(
	ctx context.Context, id int, userId int,
) (*foundationview.ContestMember, error) {
	return foundationdao.GetContestMemberDao().GetUser(ctx, id, userId)
}

func (s *ContestService) UpdateContest(
	ctx context.Context, contest *foundationmodel.Contest,
	contestProblems []*foundationmodel.ContestProblem,
	languages []string,
	authors []int,
	members []*foundationmodel.ContestMember,
	memberAuths []int,
	memberIgnores []int,
	memberVolunteers []int,
) error {
	return foundationdao.GetContestDao().UpdateContest(
		ctx,
		contest,
		contestProblems,
		languages,
		authors,
		members,
		memberAuths,
		memberIgnores,
		memberVolunteers,
	)
}

func (s *ContestService) UpdateDescription(ctx context.Context, id int, description string) error {
	return foundationdao.GetContestDao().UpdateDescription(ctx, id, description)
}

func (s *ContestService) PostPassword(ctx context.Context, userId int, contestId int, password string) (bool, error) {
	return foundationdao.GetContestDao().PostPassword(ctx, userId, contestId, password)
}

func (s *ContestService) InsertContest(
	ctx context.Context, contest *foundationmodel.Contest,
	contestProblems []*foundationmodel.ContestProblem,
	languages []string,
	authors []int,
	members []*foundationmodel.ContestMember,
	memberAuths []int,
	memberIgnores []int,
	memberVolunteers []int,
) error {
	return foundationdao.GetContestDao().InsertContest(
		ctx,
		contest,
		contestProblems,
		languages,
		authors,
		members,
		memberAuths,
		memberIgnores,
		memberVolunteers,
	)
}

func (s *ContestService) DolosContest(ctx context.Context, id int) (*string, error) {

	tempDir, err := os.MkdirTemp("", "didaoj-contest-data-*")
	if err != nil {
		return nil, metaerror.Wrap(err, "创建临时目录失败")
	}
	tempDir = filepath.ToSlash(tempDir)
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "<UNK>: "+path))
		}
	}(tempDir)

	csvContent := "filename,label,created_at,full_name\n"

	cacheNickname := make(map[int]string)

	err = foundationdao.GetJudgeJobDao().ForeachContestAcCodes(
		ctx, id, func(judgeId int, code string, problemId string, createTime time.Time, authorId int) error {
			// 保存到临时目录
			fileName := strconv.Itoa(judgeId) + ".cpp"
			filePath := path.Join(tempDir, fileName)
			err := os.WriteFile(filePath, []byte(code), 0644)
			if err != nil {
				return metaerror.Wrap(err, "写入代码到临时文件失败")
			}
			authorNickname, ok := cacheNickname[authorId]
			if !ok {
				user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, authorId)
				if err != nil {
					return metaerror.Wrap(err, "获取用户信息失败")
				}
				if user == nil {
					return metaerror.New("用户不存在")
				}
				authorNickname = user.Nickname
				cacheNickname[authorId] = authorNickname
			}
			// 添加到CSV内容
			csvContent += fileName + "," +
				problemId + "," +
				createTime.Format(time.RFC3339) + "," +
				authorNickname + "\n"
			return nil
		},
	)
	if err != nil {
		return nil, err
	}
	csvFilePath := path.Join(tempDir, "info.csv")
	err = os.WriteFile(csvFilePath, []byte(csvContent), 0644)
	if err != nil {
		return nil, metaerror.Wrap(err, "写入CSV文件失败")
	}
	// 将临时目录打包成ZIP文件
	zipName := path.Base(tempDir) + ".zip"
	err = metazip.PackagePath(tempDir, zipName)
	if err != nil {
		return nil, metaerror.Wrap(err, "打包临时目录失败")
	}
	zipFilePath := path.Join(tempDir, zipName)
	zipFile, err := os.Open(zipFilePath)
	if err != nil {
		return nil, metaerror.Wrap(err, "打开ZIP文件失败")
	}
	defer func(zipFile *os.File) {
		err := zipFile.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "关闭ZIP文件失败"))
		}
	}(zipFile)
	// 创建表单
	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)
	fileField, err := writer.CreateFormFile("dataset[zipfile]", filepath.Base(zipFilePath))
	if err != nil {
		return nil, metaerror.Wrap(err, "<UNK>")
	}
	_, err = io.Copy(fileField, zipFile)
	if err != nil {
		return nil, metaerror.Wrap(err, "复制ZIP文件内容到表单失败")
	}
	// 添加数据字段
	err = writer.WriteField("dataset[name]", metapath.GetBaseName(zipName))
	if err != nil {
		return nil, metaerror.Wrap(err, "写入表单字段失败")
	}
	err = writer.Close()
	if err != nil {
		return nil, metaerror.Wrap(err, "关闭表单失败")
	}
	req, err := http.NewRequest("POST", "https://dolos.ugent.be/api/reports", &buf)
	if err != nil {
		return nil, metaerror.Wrap(err, "创建HTTP请求失败")
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, metaerror.Wrap(err, "发送HTTP请求失败")
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "关闭响应体失败"))
		}
	}(resp.Body)
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, metaerror.Wrap(err, "读取HTTP响应失败")
	}
	type response struct {
		HTMLUrl string `json:"html_url"`
	}
	var res response
	err = json.Unmarshal(body, &res)
	if err != nil {
		return nil, metaerror.Wrap(err, "解析HTTP响应失败")
	}
	return &res.HTMLUrl, nil
}

func (s *ContestService) PostContestMemberName(
	ctx context.Context, userId int, contestId int, name string,
) error {
	return foundationdao.GetContestMemberDao().PostContestMemberName(ctx, userId, contestId, name)
}
