package foundationservice

import (
	"bytes"
	"context"
	"fmt"
	foundationauth "foundation/foundation-auth"
	"foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"log/slog"
	cfr2 "meta/cf-r2"
	metaerrorcode "meta/error-code"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metamath "meta/meta-math"
	metamd5 "meta/meta-md5"
	metapanic "meta/meta-panic"
	metapath "meta/meta-path"
	metaslice "meta/meta-slice"
	metastring "meta/meta-string"
	metazip "meta/meta-zip"
	"meta/retry"
	"meta/routine"
	"meta/singleton"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"sync"
	"time"
	weberrorcode "web/error-code"
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

func (s *ProblemService) CheckEditAuth(ctx *gin.Context, id string) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		problem, err := foundationdao.GetProblemDao().GetProblemEditAuth(ctx, id)
		if err != nil {
			return userId, false, err
		}
		if problem == nil {
			return userId, false, nil
		}
		if problem.CreatorId != userId && !slices.Contains(problem.AuthMembers, userId) {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *ProblemService) CheckSubmitAuth(ctx *gin.Context, id string) (
	int,
	bool,
	error,
) {
	userId, hasAuth, err := GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		return userId, false, err
	}
	if userId <= 0 {
		return userId, false, nil
	}
	if !hasAuth {
		problem, err := foundationdao.GetProblemDao().GetProblemViewAuth(ctx, id)
		if err != nil {
			return userId, false, err
		}
		if problem == nil {
			return userId, false, nil
		}
		if problem.Private &&
			problem.CreatorId != userId &&
			!slices.Contains(problem.Members, userId) &&
			!slices.Contains(problem.AuthMembers, userId) {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *ProblemService) GetProblemView(
	ctx context.Context,
	id string,
	userId int,
	hasAuth bool,
) (*foundationmodel.Problem, error) {
	problem, err := foundationdao.GetProblemDao().GetProblemView(ctx, id, userId, hasAuth)
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

func (s *ProblemService) GetProblemIdByContest(ctx *gin.Context, contestId int, problemIndex int) (*string, error) {
	return foundationdao.GetContestDao().GetProblemIdByContest(ctx, contestId, problemIndex)
}

func (s *ProblemService) GetProblemDescription(
	ctx context.Context,
	id string,
) (*string, error) {
	return foundationdao.GetProblemDao().GetProblemDescription(ctx, id)
}

func (s *ProblemService) GetProblemViewJudgeData(ctx context.Context, id string) (*foundationmodel.Problem, error) {
	problem, err := foundationdao.GetProblemDao().GetProblemViewJudgeData(ctx, id)
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

func (s *ProblemService) GetProblemViewApproveJudge(ctx context.Context, id string) (
	*foundationmodel.ProblemViewApproveJudge,
	error,
) {
	return foundationdao.GetProblemDao().GetProblemViewApproveJudge(ctx, id)
}

func (s *ProblemService) HasProblem(ctx context.Context, id string) (bool, error) {
	return foundationdao.GetProblemDao().HasProblem(ctx, id)
}

func (s *ProblemService) HasProblemTitle(ctx *gin.Context, title string) (bool, error) {
	return foundationdao.GetProblemDao().HasProblemTitle(ctx, title)
}

func (s *ProblemService) GetProblemList(
	ctx context.Context,
	oj string, title string, tag string,
	page int, pageSize int,
) ([]*foundationmodel.Problem, int, error) {
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
	return foundationdao.GetProblemDao().GetProblemList(
		ctx, oj, title, tags, false,
		-1, false,
		page, pageSize,
	)
}

func (s *ProblemService) GetProblemListWithUser(
	ctx context.Context, userId int, hasAuth bool,
	oj string, title string, tag string, private bool,
	page int, pageSize int,
) ([]*foundationmodel.Problem, int, map[string]foundationmodel.ProblemAttemptStatus, error) {
	var tags []int
	if tag != "" {
		var err error
		tags, err = foundationdao.GetProblemTagDao().SearchTags(ctx, tag)
		if err != nil {
			return nil, 0, nil, err
		}
		if len(tags) == 0 {
			return nil, 0, nil, nil
		}
	}
	problemList, totalCount, err := foundationdao.GetProblemDao().GetProblemList(
		ctx, oj, title, tags, private,
		userId, hasAuth,
		page, pageSize,
	)
	if err != nil {
		return nil, 0, nil, err
	}
	if len(problemList) <= 0 {
		return nil, 0, nil, nil
	}
	var problemIds []string
	for _, problem := range problemList {
		problemIds = append(problemIds, problem.Id)
	}
	problemStatus, err := foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(ctx, problemIds, userId, -1, nil, nil)
	if err != nil {
		return nil, 0, nil, err
	}
	return problemList, totalCount, problemStatus, nil
}

func (s *ProblemService) GetProblemRecommend(
	ctx context.Context,
	userId int,
	hasAuth bool,
	problemId string,
) ([]*foundationmodel.Problem, error) {
	var err error
	var problemIds []string
	if problemId == "" {
		problemIds, err = foundationdao.GetJudgeJobDao().GetProblemRecommendByUser(ctx, userId, hasAuth)
	} else {
		problemIds, err = foundationdao.GetJudgeJobDao().GetProblemRecommendByProblem(ctx, userId, hasAuth, problemId)
	}
	if err != nil {
		return nil, err
	}
	if len(problemIds) == 0 {
		return nil, nil
	}
	sort.Slice(
		problemIds, func(a, b int) bool {
			lengthA := len(problemIds[a])
			lengthB := len(problemIds[b])
			if lengthA != lengthB {
				return lengthA < lengthB
			}
			return strings.Compare(problemIds[a], problemIds[b]) < 0
		},
	)
	problemList, err := foundationdao.GetProblemDao().GetProblems(ctx, problemIds)
	if err != nil {
		return nil, err
	}
	if len(problemList) == 0 {
		return nil, nil
	}
	return problemList, nil
}

func (s *ProblemService) GetProblemTagList(ctx context.Context, maxCount int) (
	[]*foundationmodel.ProblemTag,
	int,
	error,
) {
	return foundationdao.GetProblemTagDao().GetProblemTagList(ctx, maxCount)
}

func (s *ProblemService) GetProblemTagByIds(ctx context.Context, ids []int) ([]*foundationmodel.ProblemTag, error) {
	return foundationdao.GetProblemTagDao().GetProblemTagByIds(ctx, ids)
}

func (s *ProblemService) GetProblemTitles(ctx *gin.Context, userId int, hasAuth bool, problems []string) (
	[]*foundationmodel.ProblemViewTitle,
	error,
) {
	return foundationdao.GetProblemDao().GetProblemTitles(ctx, userId, hasAuth, problems)
}

func (s *ProblemService) FilterValidProblemIds(ctx *gin.Context, ids []string) ([]string, error) {
	return foundationdao.GetProblemDao().FilterValidProblemIds(ctx, ids)
}

func (s *ProblemService) InsertProblem(
	ctx context.Context,
	problem *foundationmodel.Problem,
	tags []string,
) (*string, error) {
	return foundationdao.GetProblemDao().PostCreate(ctx, problem, tags)
}

func (s *ProblemService) UpdateProblem(
	ctx context.Context,
	problemId string,
	problem *foundationmodel.Problem,
	tags []string,
) error {
	return foundationdao.GetProblemDao().UpdateProblem(ctx, problemId, problem, tags)
}

func (s *ProblemService) PostJudgeData(
	ctx context.Context,
	problemId string,
	unzipDir string,
	oldMd5 *string,
	goJudgeUrl string,
	checkR2FileCount bool,
) error {
	// 如果包含文件夹，认为失败
	err := filepath.Walk(
		unzipDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				// 跳过根目录本身
				if path != unzipDir {
					return metaerror.New("<UNK>: " + path + " is not a directory")
				}
				return nil
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataCannotDir)
	}

	judgeType := foundationjudge.JudgeTypeNormal

	var jobConfig foundationjudge.JudgeJobConfig

	// 解析rule.yaml
	ruleFile := filepath.Join(unzipDir, "rule.yaml")
	yamlFile, err := os.ReadFile(ruleFile)
	if err == nil {
		err = yaml.Unmarshal(yamlFile, &jobConfig)
		if err != nil {
			return metaerror.NewCode(weberrorcode.ProblemJudgeDataRuleYamlFail)
		}
	} else {
		if !os.IsNotExist(err) {
			return metaerror.NewCode(metaerrorcode.CommonError)
		}
	}

	if jobConfig.SpecialJudge == nil {
		specialFiles := map[string]string{
			"spj.c":   "c",
			"spj.cc":  "cpp",
			"spj.cpp": "cpp",
		}
		// 判断是否存在对应文件
		for fileName, language := range specialFiles {
			filePath := path.Join(unzipDir, fileName)
			_, err := os.Stat(filePath)
			if err == nil {
				jobConfig.SpecialJudge = &foundationjudge.SpecialJudgeConfig{}
				jobConfig.SpecialJudge.Language = language
				jobConfig.SpecialJudge.Source = fileName
				break
			}
		}
	}

	if jobConfig.SpecialJudge != nil {
		runUrl := metahttp.UrlJoin(goJudgeUrl, "run")

		language := foundationjudge.GetLanguageByKey(jobConfig.SpecialJudge.Language)
		if !foundationjudge.IsValidJudgeLanguage(int(language)) {
			return metaerror.NewCode(weberrorcode.ProblemJudgeDataSpjLanguageNotValid)
		}

		// 考虑编译机性能影响，暂时仅允许部分语言
		if !foundationjudge.IsValidSpecialJudgeLanguage(language) {
			return metaerror.NewCode(weberrorcode.ProblemJudgeDataSpjLanguageNotValid)
		}

		codeFilePath := filepath.Join(unzipDir, jobConfig.SpecialJudge.Source)
		codeContent, err := metastring.GetStringFromOpenFile(codeFilePath)
		if err != nil {
			return metaerror.NewCode(weberrorcode.ProblemJudgeDataSpjContentNotValid)

		}

		jobKey := uuid.New().String()

		execFileIds, extraMessage, compileStatus, err := foundationjudge.CompileCode(
			jobKey,
			runUrl,
			language,
			codeContent,
			nil,
		)
		if extraMessage != "" {
			slog.Warn("judge compile", "extraMessage", extraMessage, "compileStatus", compileStatus)
		}
		if compileStatus != foundationjudge.JudgeStatusAC {
			return metaerror.NewCode(weberrorcode.ProblemJudgeDataSpjCompileFail)
		}
		if err != nil {
			metapanic.ProcessError(err)
			return metaerror.NewCode(weberrorcode.ProblemJudgeDataSpjCompileFail)
		}
		for _, fileId := range execFileIds {
			deleteUrl := metahttp.UrlJoin(goJudgeUrl, "file", fileId)
			err := foundationjudge.DeleteFile(jobKey, deleteUrl)
			if err != nil {
				metapanic.ProcessError(err)
			}
		}
		judgeType = foundationjudge.JudgeTypeSpecial
	}

	if len(jobConfig.Tasks) <= 0 {
		// 如果没有rule.yaml文件，则根据文件生成Config信息
		files, err := os.ReadDir(unzipDir)
		if err != nil {
			return metaerror.NewCode(metaerrorcode.CommonError)
		}
		taskKeyMap := make(map[string]bool)
		hasInFiles := make(map[string]bool)
		hasOutFiles := make(map[string]bool)
		for _, file := range files {
			fileBaseName := metapath.GetBaseName(file.Name())
			if strings.HasSuffix(file.Name(), ".out") {
				hasOutFiles[fileBaseName] = true
			} else if strings.HasSuffix(file.Name(), ".in") {
				hasInFiles[fileBaseName] = true
			}
			taskKeyMap[fileBaseName] = true
		}
		var taskKeys []string
		for key, _ := range taskKeyMap {
			taskKeys = append(taskKeys, key)
		}
		taskKeys = metaslice.RemoveAllFunc(
			taskKeys, func(key string) bool {
				return !hasInFiles[key] && !hasOutFiles[key]
			},
		)
		taskCount := len(taskKeys)

		if taskCount <= 0 {
			return metaerror.NewCode(weberrorcode.ProblemJudgeDataWithoutTask)
		}

		sort.Slice(
			taskKeys, func(i, j int) bool {
				return taskKeys[i] < taskKeys[j]
			},
		)

		for _, key := range taskKeys {
			if !hasInFiles[key] && !hasOutFiles[key] {
				continue
			}
			judgeTaskConfig := &foundationjudge.JudgeTaskConfig{
				Key: key,
			}
			if hasInFiles[key] {
				judgeTaskConfig.InFile = key + ".in"
			}
			if hasOutFiles[key] {
				judgeTaskConfig.OutFile = key + ".out"
				outFile, err := os.Stat(path.Join(unzipDir, judgeTaskConfig.OutFile))
				if err != nil {
					return metaerror.NewCode(weberrorcode.ProblemJudgeDataTaskLoadFail)
				}
				judgeTaskConfig.OutLimit = metamath.Max(outFile.Size()*2, 1024)
			} else {
				// 考虑到SpecialJudge的情况可能也需要输出，这里默认给个大小
				if jobConfig.SpecialJudge != nil {
					judgeTaskConfig.OutLimit = 1048576 * 1 //1MB
				}
			}
			jobConfig.Tasks = append(jobConfig.Tasks, judgeTaskConfig)
		}
	}

	taskCount := len(jobConfig.Tasks)

	if taskCount <= 0 {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataWithoutTask)
	}

	if taskCount > 200 {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataTaskCountTooMany1000)
	}

	totalScore := 0
	for _, taskConfig := range jobConfig.Tasks {
		totalScore += taskConfig.Score
	}
	leftScore := 0
	if totalScore <= 0 {
		totalScore = 1000
		averageScore := totalScore / taskCount
		for _, taskConfig := range jobConfig.Tasks {
			taskConfig.Score = averageScore
		}
		leftScore = totalScore % taskCount
	} else {
		//把totalScore转为0~1000
		rate := 1000.0 / float64(totalScore)
		totalScore = 1000
		sumScore := 0
		for _, taskConfig := range jobConfig.Tasks {
			taskConfig.Score = int(float64(taskConfig.Score) * rate)
			sumScore += taskConfig.Score
		}
		leftScore = totalScore - sumScore
	}
	for i := taskCount - 1; i >= 0 && leftScore > 0; i-- {
		jobConfig.Tasks[i].Score += 1
		leftScore--
	}

	// 重新生成一个rule.yaml
	ruleFile = filepath.Join(unzipDir, "rule.yaml")
	yamlData, err := yaml.Marshal(jobConfig)
	if err != nil {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataRuleYamlFail)
	}
	err = os.WriteFile(ruleFile, yamlData, 0644)
	if err != nil {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataRuleYamlFail)
	}

	// 把所有文件的换行改为Linux格式
	err = filepath.Walk(
		unzipDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			content, err := os.ReadFile(path)
			if err != nil {
				return metaerror.Wrap(err, "<UNK>: "+path+" is not readable")
			}
			// 将 CRLF (\r\n) 和 CR (\r) 替换为 LF (\n)
			normalized := bytes.ReplaceAll(content, []byte("\r\n"), []byte("\n"))
			normalized = bytes.ReplaceAll(normalized, []byte("\r"), []byte("\n"))
			// 写回文件
			err = os.WriteFile(path, normalized, 0644)
			if err != nil {
				return fmt.Errorf("写入文件失败: %s, %w", path, err)
			}
			return nil
		},
	)
	if err != nil {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataProcessWrapLineFail)
	}

	var uploadFiles []string
	err = filepath.Walk(
		unzipDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			uploadFiles = append(uploadFiles, path)
			return nil
		},
	)

	judgeDataMd5, err := metamd5.MultiFileMD5(uploadFiles)
	if err != nil {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataProcessMd5Fail)
	}
	slog.Info("judge data md5", "problemId", problemId, "md5", judgeDataMd5)

	// 上传r2
	r2Client := cfr2.GetSubsystem().GetClient("judge-data")
	if r2Client == nil {
		return metaerror.NewCode(metaerrorcode.CommonError)
	}

	if oldMd5 != nil && *oldMd5 == judgeDataMd5 {
		if checkR2FileCount {
			var oldKeys []string
			input := &s3.ListObjectsV2Input{
				Bucket: aws.String("didaoj-judge"),
				Prefix: aws.String(path.Join(problemId, *oldMd5)),
			}
			err = r2Client.ListObjectsV2PagesWithContext(
				ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
					for _, obj := range page.Contents {
						oldKeys = append(oldKeys, *obj.Key)
					}
					return true
				},
			)
			// 正常情况下R2中应该存在所有文件+1个汇总的压缩包
			if len(uploadFiles)+1 == len(oldKeys) {
				return nil
			}
		} else {
			return nil
		}
	}

	var maxConcurrency = 10

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxConcurrency)
	errChan := make(chan error, 1) // 只保留第一个错误
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	for _, uploadFile := range uploadFiles {
		select {
		case <-ctx.Done():
			break
		default:
		}
		wg.Add(1)
		sem <- struct{}{} // acquire
		goPath := uploadFile
		routine.SafeGo(
			"upload judge data file", func() error {
				defer wg.Done()
				defer func() { <-sem }() // release
				relativePath, err := filepath.Rel(unzipDir, goPath)
				if err != nil {
					select {
					case errChan <- err:
					default:
					}
					return nil
				}
				key := filepath.ToSlash(filepath.Join(problemId, judgeDataMd5, relativePath))
				file, err := os.Open(goPath)
				if err != nil {
					select {
					case errChan <- err:
					default:
					}
					return nil
				}
				defer func() {
					if err := file.Close(); err != nil {
						metapanic.ProcessError(metaerror.Wrap(err, "close file error"))
					}
				}()
				slog.Info("put object start", "key", key)
				var finalErr error
				err = retry.TryRetrySleep(
					"put object", 3, time.Second*3, func(i int) bool {
						_, err = r2Client.PutObjectWithContext(
							ctx, &s3.PutObjectInput{
								Bucket: aws.String("didaoj-judge"),
								Key:    aws.String(key),
								Body:   file,
							},
						)
						if err != nil {
							finalErr = err
							return true
						}
						return true
					},
				)
				if err != nil {
					slog.Info("put object error", "key", key)
					select {
					case errChan <- metaerror.Wrap(finalErr, "put object error, key:%s", key):
					default:
					}
					return nil
				}
				slog.Info("put object success", "key", key)
				return nil
			},
		)
	}

	// 等待所有任务完成
	wg.Wait()
	close(errChan)

	if err, ok := <-errChan; ok {
		metapanic.ProcessError(err)
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataSubmitFail)
	}

	zipFileName := fmt.Sprintf("%s-%s.zip", problemId, judgeDataMd5)
	err = metazip.PackagePath(unzipDir, zipFileName)
	if err != nil {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataSubmitFail)
	}
	zipFile, err := os.Open(path.Join(unzipDir, zipFileName))
	defer func() {
		if err := zipFile.Close(); err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "close zip file error"))
		}
	}()
	zipKey := filepath.ToSlash(filepath.Join(problemId, judgeDataMd5, zipFileName))
	_, err = r2Client.PutObjectWithContext(
		ctx, &s3.PutObjectInput{
			Bucket: aws.String("didaoj-judge"),
			Key:    aws.String(zipKey),
			Body:   zipFile,
		},
	)
	if err != nil {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataSubmitFail)
	}

	err = s.UpdateProblemJudgeInfo(ctx, problemId, judgeType, judgeDataMd5)
	if err != nil {
		return err
	}

	// 删除旧的路径
	putPrefix := filepath.ToSlash(path.Join(problemId, judgeDataMd5))
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("didaoj-judge"),
		Prefix: aws.String(problemId),
	}
	var deleteKeys []string
	err = r2Client.ListObjectsV2PagesWithContext(
		ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				if strings.HasPrefix(*obj.Key, putPrefix) {
					continue
				}
				deleteKeys = append(deleteKeys, *obj.Key)
			}
			return true
		},
	)

	if len(deleteKeys) > 0 {
		// 这里不应该依赖PostJudge的流程，走到这里必须执行完毕删除逻辑，否则就会产生遗留数据，只能等待下一次更新判题数据时删除
		routine.SafeGo(
			"delete judge data object", func() error {
				sem = make(chan struct{}, maxConcurrency)
				errChan = make(chan error, 1) // 只保留第一个错误
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				for _, delKey := range deleteKeys {
					select {
					case <-ctx.Done():
						break
					default:
					}
					wg.Add(1)
					sem <- struct{}{}    // acquire
					currentKey := delKey // 避免闭包问题
					routine.SafeGo(
						"delete judge data object", func() error {
							defer wg.Done()
							defer func() { <-sem }() // release

							slog.Info("delete object start", "key", currentKey)

							var finalErr error
							err := retry.TryRetrySleep(
								"delete object", 3, time.Second*3, func(i int) bool {
									_, err := r2Client.DeleteObjectWithContext(
										ctx, &s3.DeleteObjectInput{
											Bucket: aws.String("didaoj-judge"),
											Key:    aws.String(currentKey),
										},
									)
									if err != nil {
										finalErr = err
										return true // 重试
									}
									return true // 成功也退出
								},
							)
							if err != nil {
								slog.Info("delete object error", "key", currentKey)
								select {
								case errChan <- metaerror.Wrap(finalErr, "delete object error, key:%s", currentKey):
								default:
								}
								return nil
							}
							slog.Info("delete object success", "key", currentKey)
							return nil
						},
					)
				}

				// 等待所有任务完成
				wg.Wait()
				close(errChan)

				if err, ok := <-errChan; ok {
					metapanic.ProcessError(err)
					return metaerror.NewCode(weberrorcode.ProblemJudgeDataSubmitFail)
				}
				return nil
			},
		)
	}

	return nil
}

func (s *ProblemService) PostDailyCreate(
	ctx *gin.Context,
	problemDaily *foundationmodel.ProblemDaily,
) error {
	return foundationdao.GetProblemDailyDao().PostDailyCreate(ctx, problemDaily)
}

func (s *ProblemService) PostDailyEdit(
	ctx *gin.Context,
	id string,
	problemDaily *foundationmodel.ProblemDaily,
) error {
	return foundationdao.GetProblemDailyDao().UpdateProblemDaily(ctx, id, problemDaily)
}

func (s *ProblemService) UpdateProblemDescription(
	ctx context.Context,
	id string,
	description string,
) error {
	return foundationdao.GetProblemDao().UpdateProblemDescription(ctx, id, description)
}

func (s *ProblemService) UpdateProblemJudgeInfo(
	ctx context.Context,
	id string,
	judgeType foundationjudge.JudgeType,
	md5 string,
) error {
	return foundationdao.GetProblemDao().UpdateProblemJudgeInfo(ctx, id, judgeType, md5)
}

func (s *ProblemService) HasProblemDaily(ctx *gin.Context, dailyId string) (bool, error) {
	return foundationdao.GetProblemDailyDao().HasProblemDaily(ctx, dailyId)
}

func (s *ProblemService) HasProblemDailyProblem(ctx *gin.Context, problemId string) (bool, error) {
	return foundationdao.GetProblemDailyDao().HasProblemDailyProblem(ctx, problemId)
}

func (s *ProblemService) GetProblemIdByDaily(ctx *gin.Context, dailyId string, hasAuth bool) (*string, error) {
	return foundationdao.GetProblemDailyDao().GetProblemIdByDaily(ctx, dailyId, hasAuth)
}

func (s *ProblemService) GetProblemDaily(ctx *gin.Context, dailyId string, hasAuth bool) (
	*foundationmodel.ProblemDaily,
	error,
) {
	return foundationdao.GetProblemDailyDao().GetProblemDaily(ctx, dailyId, hasAuth)
}

func (s *ProblemService) GetProblemDailyEdit(ctx *gin.Context, dailyId string) (*foundationmodel.ProblemDaily, error) {
	daily, err := foundationdao.GetProblemDailyDao().GetProblemDailyEdit(ctx, dailyId)
	if err != nil {
		return nil, err
	}
	if daily == nil {
		return nil, nil
	}
	if daily.CreatorId > 0 {
		user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, daily.CreatorId)
		if err != nil {
			return nil, err
		}
		if user == nil {
			return nil, nil
		}
		daily.CreatorUsername = &user.Username
		daily.CreatorNickname = &user.Nickname
	}
	if daily.UpdaterId > 0 {
		if daily.UpdaterId == daily.CreatorId {
			daily.UpdaterUsername = daily.CreatorUsername
			daily.UpdaterNickname = daily.CreatorNickname
		} else {
			user, err := foundationdao.GetUserDao().GetUserAccountInfo(ctx, daily.UpdaterId)
			if err != nil {
				return nil, err
			}
			if user == nil {
				return nil, nil
			}
			daily.UpdaterUsername = &user.Username
			daily.UpdaterNickname = &user.Nickname
		}
	}
	return daily, nil
}

func (s *ProblemService) GetDailyList(
	ctx *gin.Context,
	userId int,
	hasAuth bool,
	startDate *string,
	endDate *string,
	problemId string,
	page int,
	pageSize int,
) (
	[]*foundationmodel.ProblemDaily,
	int,
	[]*foundationmodel.ProblemTag,
	map[string]foundationmodel.ProblemAttemptStatus,
	error,
) {
	dailyList, totalCount, err := foundationdao.GetProblemDailyDao().GetDailyList(
		ctx,
		hasAuth,
		startDate,
		endDate,
		problemId,
		page,
		pageSize,
	)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	if len(dailyList) == 0 {
		return nil, 0, nil, nil, nil
	}
	var problemIds []string
	for _, daily := range dailyList {
		problemIds = append(problemIds, daily.ProblemId)
	}
	problemList, err := foundationdao.GetProblemDao().GetProblems(ctx, problemIds)
	if err != nil {
		return nil, 0, nil, nil, err
	}
	var tagIds []int
	for _, problem := range problemList {
		tagIds = append(tagIds, problem.Tags...)
	}
	var tags []*foundationmodel.ProblemTag
	if len(tagIds) > 0 {
		tags, err = foundationdao.GetProblemTagDao().GetProblemTagByIds(ctx, tagIds)
		if err != nil {
			return nil, 0, nil, nil, err
		}
	}
	var problemStatus map[string]foundationmodel.ProblemAttemptStatus
	if userId > 0 {
		problemStatus, err = foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
			ctx,
			problemIds,
			userId,
			-1,
			nil,
			nil,
		)
		if err != nil {
			return nil, 0, nil, nil, err
		}
	}
	problemMap := make(map[string]*foundationmodel.Problem)
	for _, problem := range problemList {
		problemMap[problem.Id] = problem
	}
	for _, daily := range dailyList {
		problem, ok := problemMap[daily.ProblemId]
		if ok {
			daily.Title = &problem.Title
			daily.Tags = problem.Tags
			daily.Accept = problem.Accept
			daily.Attempt = problem.Attempt
		}
	}
	return dailyList, totalCount, tags, problemStatus, nil
}

func (s *ProblemService) GetDailyRecently(ctx *gin.Context, userId int) (
	[]*foundationmodel.ProblemDaily,
	map[string]foundationmodel.ProblemAttemptStatus,
	error,
) {
	daily, err := foundationdao.GetProblemDailyDao().GetDailyRecently(ctx)
	if err != nil {
		return nil, nil, err
	}
	if daily == nil {
		return nil, nil, nil
	}
	for _, d := range daily {
		title, err := foundationdao.GetProblemDao().GetProblemTitle(ctx, &d.ProblemId)
		if err == nil {
			d.Title = title
		} else {
			titlePtr := "未知题目"
			d.Title = &titlePtr
		}
	}
	var problemAttemptStatus map[string]foundationmodel.ProblemAttemptStatus
	if userId > 0 {
		problemIds := make([]string, len(daily))
		for i, d := range daily {
			problemIds[i] = d.ProblemId
		}
		problemAttemptStatus, err = foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
			ctx, problemIds, userId, -1, nil, nil,
		)
		if err != nil {
			return nil, nil, err
		}
	}
	return daily, problemAttemptStatus, nil
}
