package foundationservice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	foundationauth "foundation/foundation-auth"
	foundationdao "foundation/foundation-dao"
	foundationenum "foundation/foundation-enum"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"io"
	"log/slog"
	cfr2 "meta/cf-r2"
	metaerrorcode "meta/error-code"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metamath "meta/meta-math"
	metamd5 "meta/meta-md5"
	metapanic "meta/meta-panic"
	metapath "meta/meta-path"
	metaredis "meta/meta-redis"
	metaslice "meta/meta-slice"
	metastring "meta/meta-string"
	metazip "meta/meta-zip"
	"meta/routine"
	"meta/singleton"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
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

func (s *ProblemService) CheckEditAuth(ctx *gin.Context, id int) (
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
		hasAuth, err = foundationdao.GetProblemDao().CheckProblemEditAuth(ctx, id, userId)
		if err != nil {
			return userId, false, err
		}
		if !hasAuth {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *ProblemService) CheckSubmitAuth(ctx *gin.Context, id int) (
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
		hasAuth, err = foundationdao.GetProblemDao().CheckProblemSubmitAuth(ctx, id, userId)
		if err != nil {
			return userId, false, err
		}
		if !hasAuth {
			return userId, false, nil
		}
	}
	return userId, true, nil
}

func (s *ProblemService) GetProblemView(
	ctx context.Context,
	id int,
	userId int,
	hasAuth bool,
) (*foundationview.Problem, error) {
	return foundationdao.GetProblemDao().GetProblemView(ctx, id, userId, hasAuth)
}

func (s *ProblemService) GetProblemTags(ctx context.Context, id int) ([]*foundationmodel.Tag, error) {
	return s.GetProblemsTags(ctx, []int{id})
}

func (s *ProblemService) GetProblemsTags(ctx context.Context, ids []int) ([]*foundationmodel.Tag, error) {
	tagIds, err := foundationdao.GetProblemTagDao().GetProblemTags(ctx, ids)
	if err != nil {
		return nil, err
	}
	tags, err := foundationdao.GetTagDao().GetTags(ctx, tagIds)
	if err != nil {
		return nil, err
	}
	tagLen := len(tags)
	if tagLen == 0 {
		return nil, nil
	}
	if tagLen == 1 {
		return tags, nil
	}
	tagMap := make(map[int]*foundationmodel.Tag)
	for _, tag := range tags {
		tagMap[tag.Id] = tag
	}
	var resultTags []*foundationmodel.Tag
	for _, tagId := range tagIds {
		if tag, ok := tagMap[tagId]; ok {
			resultTags = append(resultTags, tag)
		}
	}
	return resultTags, nil
}

func (s *ProblemService) GetProblemIdByContest(ctx *gin.Context, contestId int, problemIndex int) (int, error) {
	return foundationdao.GetContestProblemDao().GetProblemId(ctx, contestId, problemIndex)
}

func (s *ProblemService) GetProblemDescription(
	ctx context.Context,
	id int,
) (*string, error) {
	return foundationdao.GetProblemDao().GetProblemDescription(ctx, id)
}

func (s *ProblemService) GetProblemViewJudgeData(ctx context.Context, id int) (
	*foundationview.ProblemJudgeData,
	error,
) {
	return foundationdao.GetProblemDao().GetProblemViewJudgeData(ctx, id)
}

func (s *ProblemService) GetProblemViewApproveJudge(ctx context.Context, id int) (
	*foundationview.ProblemViewApproveJudge,
	error,
) {
	return foundationdao.GetProblemDao().GetProblemViewApproveJudge(ctx, id)
}

func (s *ProblemService) HasProblem(ctx context.Context, id int) (bool, error) {
	return foundationdao.GetProblemDao().HasProblem(ctx, id)
}

func (s *ProblemService) HasProblemByKey(ctx context.Context, key string) (bool, error) {
	return foundationdao.GetProblemDao().HasProblemByKey(ctx, key)
}

func (s *ProblemService) HasProblemTitle(ctx *gin.Context, title string) (bool, error) {
	return foundationdao.GetProblemDao().HasProblemTitle(ctx, title)
}

func (s *ProblemService) GetProblemIdByKey(ctx context.Context, problemKey string) (int, error) {
	return foundationdao.GetProblemDao().GetProblemIdByKey(ctx, problemKey)
}
func (s *ProblemService) GetProblemIdsByKey(ctx context.Context, problemKeys []string) ([]int, error) {
	return foundationdao.GetProblemDao().GetProblemIdsByKey(ctx, problemKeys)
}

func (s *ProblemService) FilterValidProblemIds(ctx context.Context, ids []int) ([]int, error) {
	return foundationdao.GetProblemDao().FilterValidProblemIds(ctx, ids)
}

func (s *ProblemService) GetProblemList(
	ctx context.Context,
	oj string, title string, tag string,
	page int, pageSize int,
) ([]*foundationview.ProblemViewList, int, error) {
	var searchTags []int
	if tag != "" {
		var err error
		searchTags, err = foundationdao.GetTagDao().SearchTagIds(ctx, tag)
		if err != nil {
			return nil, 0, err
		}
		if len(searchTags) == 0 {
			return nil, 0, nil
		}
	}
	list, totalCount, err := foundationdao.GetProblemDao().GetProblemList(
		ctx, oj, title, searchTags, false,
		-1, false,
		page, pageSize,
	)
	if err != nil {
		return nil, 0, err
	}
	var problemIds []int
	for _, problem := range list {
		problemIds = append(problemIds, problem.Id)
	}
	problemMap, err := foundationdao.GetProblemTagDao().GetProblemTagMap(ctx, problemIds)
	if err != nil {
		return nil, 0, err
	}
	for _, problem := range list {
		if tags, ok := problemMap[problem.Id]; ok {
			problem.Tags = tags
		}
	}
	return list, totalCount, err
}

func (s *ProblemService) GetProblemListWithUser(
	ctx context.Context, userId int, hasAuth bool,
	oj string, title string, tag string, private bool,
	page int, pageSize int,
) ([]*foundationview.ProblemViewList, int, map[int]foundationenum.ProblemAttemptStatus, error) {
	var tags []int
	if tag != "" {
		var err error
		tags, err = foundationdao.GetTagDao().SearchTagIds(ctx, tag)
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
	var problemIds []int
	for _, problem := range problemList {
		problemIds = append(problemIds, problem.Id)
	}
	problemStatus, err := foundationdao.GetJudgeJobDao().GetProblemAttemptStatus(
		ctx,
		userId,
		problemIds,
		-1,
		nil,
		nil,
	)
	if err != nil {
		return nil, 0, nil, err
	}
	problemMap, err := foundationdao.GetProblemTagDao().GetProblemTagMap(ctx, problemIds)
	if err != nil {
		return nil, 0, nil, err
	}
	for _, problem := range problemList {
		if tags, ok := problemMap[problem.Id]; ok {
			problem.Tags = tags
		}
	}
	return problemList, totalCount, problemStatus, nil
}

func (s *ProblemService) GetProblemRecommend(
	ctx context.Context,
	userId int,
	hasAuth bool,
	problemId int,
) ([]*foundationview.ProblemViewList, error) {
	redisKey := fmt.Sprintf("problem_recommend_%d", userId)
	if problemId > 0 {
		redisKey += "_" + strconv.Itoa(problemId)
	}
	redisClient := metaredis.GetSubsystem().GetClient()
	cached, err := redisClient.Get(ctx, redisKey).Result()
	if err == nil && cached != "" {
		var statics []*foundationview.ProblemViewList
		if err := json.Unmarshal([]byte(cached), &statics); err == nil {
			return statics, nil
		}
	}
	problemIds, err := foundationdao.GetJudgeJobDao().GetProblemRecommendByProblem(
		ctx,
		userId,
		hasAuth,
		problemId,
	)
	if err != nil {
		return nil, err
	}
	if len(problemIds) == 0 {
		err := redisClient.Set(ctx, redisKey, "[]", time.Hour).Err()
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	problemList, err := foundationdao.GetProblemDao().SelectProblemViewList(ctx, problemIds, true)
	if err != nil {
		return nil, err
	}
	if len(problemList) == 0 {
		err := redisClient.Set(ctx, redisKey, "[]", time.Hour).Err()
		if err != nil {
			return nil, err
		}
		return nil, nil
	}
	jsonString, err := json.Marshal(problemList)
	if err != nil {
		return nil, metaerror.Wrap(err, "marshal problem list error")
	}
	err = redisClient.Set(ctx, redisKey, jsonString, time.Hour).Err()
	if err != nil {
		return nil, err
	}
	return problemList, nil
}

func (s *ProblemService) GetProblemTagList(ctx context.Context, maxCount int) (
	[]*foundationmodel.Tag,
	int,
	error,
) {
	return foundationdao.GetProblemTagDao().GetProblemTagList(ctx, maxCount)
}

func (s *ProblemService) GetProblemTitles(ctx *gin.Context, userId int, hasAuth bool, problems []int) (
	[]*foundationview.ProblemViewTitle,
	error,
) {
	return foundationdao.GetProblemDao().GetProblemTitles(ctx, userId, hasAuth, problems)
}

func (s *ProblemService) InsertProblemLocal(
	ctx context.Context,
	problem *foundationmodel.Problem,
	problemLocal *foundationmodel.ProblemLocal,
	tags []string,
) error {
	return foundationdao.GetProblemDao().InsertProblemLocal(ctx, problem, problemLocal, tags)
}

func (s *ProblemService) UpdateProblem(
	ctx context.Context,
	problemId int,
	problem *foundationmodel.Problem,
	tags []string,
) error {
	return foundationdao.GetProblemDao().UpdateProblem(ctx, problemId, problem, tags)
}

func (s *ProblemService) PostJudgeData(
	ctx context.Context,
	problemId int,
	unzipDir string,
	oldMd5 *string,
	goJudgeUrl string,
	goJudgeConfigFiles map[string]string,
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

	err = filepath.Walk(
		unzipDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() {
				return nil
			}
			if strings.HasSuffix(info.Name(), ".in") {
				return nil
			}
			if strings.HasSuffix(info.Name(), ".out") {
				return nil
			}
			if info.Name() == "rule.yaml" {
				return nil
			}
			if info.Name() == "spj.c" || info.Name() == "spj.cc" || info.Name() == "spj.cpp" {
				return nil
			}
			return metaerror.New("<UNK>: " + path + " is not a valid judge data file")
		},
	)
	if err != nil {
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataHasNotValid)
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

		goJudgeClient := http.DefaultClient

		execFileIds, extraMessage, compileStatus, err := foundationjudge.CompileCode(
			goJudgeClient,
			jobKey,
			runUrl,
			language,
			codeContent,
			goJudgeConfigFiles,
			true,
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
			err := foundationjudge.DeleteFile(goJudgeClient, jobKey, deleteUrl)
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
				Prefix: aws.String(path.Join(strconv.Itoa(problemId), *oldMd5)),
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
	errChan := make(chan error, 1)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	for _, uploadFile := range uploadFiles {
		// 如果已经有错误，取消后续任务
		select {
		case <-ctx.Done():
			break
		default:
		}

		wg.Add(1)
		sem <- struct{}{} // acquire
		goPath := uploadFile

		go func(path string) {
			defer wg.Done()
			defer func() { <-sem }() // release

			relativePath, err := filepath.Rel(unzipDir, path)
			if err != nil {
				select {
				case errChan <- err:
				default:
				}
				cancel()
				return
			}

			key := filepath.ToSlash(filepath.Join(strconv.Itoa(problemId), judgeDataMd5, relativePath))

			// 多次尝试打开文件（理论上文件打开失败重试意义不大，但为防止临时FS问题）
			var file *os.File
			for i := 0; i < 3; i++ {
				file, err = os.Open(path)
				if err == nil {
					break
				}
				time.Sleep(2 * time.Second)
			}
			if err != nil {
				select {
				case errChan <- fmt.Errorf("open file error (%s): %w", path, err):
				default:
				}
				cancel()
				return
			}
			defer func(file *os.File) {
				err := file.Close()
				if err != nil {
					metapanic.ProcessError(metaerror.Wrap(err, "close file error"))
				}
			}(file)

			slog.Info("put object start", "key", key)

			var uploadErr error
			for i := 0; i < 3; i++ {
				// 重置文件偏移，否则重试时 Body 是空的
				if _, err := file.Seek(0, io.SeekStart); err != nil {
					uploadErr = fmt.Errorf("seek file error (%s): %w", path, err)
					break
				}
				_, uploadErr = r2Client.PutObjectWithContext(
					ctx, &s3.PutObjectInput{
						Bucket: aws.String("didaoj-judge"),
						Key:    aws.String(key),
						Body:   file,
					},
				)
				if uploadErr == nil {
					break // success
				}
				slog.Warn("put object retry", "attempt", i+1, "key", key, "error", uploadErr)
				time.Sleep(3 * time.Second)
			}

			if uploadErr != nil {
				slog.Error("put object failed", "key", key, "error", uploadErr)
				select {
				case errChan <- uploadErr:
				default:
				}
				cancel()
				return
			}

			slog.Info("put object success", "key", key)
		}(goPath)
	}

	wg.Wait()
	close(errChan)

	if err, ok := <-errChan; ok {
		metapanic.ProcessError(err)
		return metaerror.NewCode(weberrorcode.ProblemJudgeDataSubmitFail)
	}

	zipFileName := fmt.Sprintf("%d-%s.zip", problemId, judgeDataMd5)
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
	zipKey := filepath.ToSlash(filepath.Join(strconv.Itoa(problemId), judgeDataMd5, zipFileName))
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
	putPrefix := filepath.ToSlash(path.Join(strconv.Itoa(problemId), judgeDataMd5))
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("didaoj-judge"),
		Prefix: aws.String(strconv.Itoa(problemId)),
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
		routine.SafeGo(
			"delete judge data object", func() error {
				var wg sync.WaitGroup
				sem := make(chan struct{}, maxConcurrency)
				errChan := make(chan error, 1)
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()

				for _, delKey := range deleteKeys {
					// 如果已出错，则终止派发新任务
					select {
					case <-ctx.Done():
						break
					default:
					}

					wg.Add(1)
					sem <- struct{}{}
					currentKey := delKey // 闭包安全

					go func(key string) {
						defer wg.Done()
						defer func() { <-sem }()

						slog.Info("delete object start", "key", key)

						var deleteErr error
						for i := 0; i < 3; i++ {
							_, err := r2Client.DeleteObjectWithContext(
								ctx, &s3.DeleteObjectInput{
									Bucket: aws.String("didaoj-judge"),
									Key:    aws.String(key),
								},
							)
							if err == nil {
								slog.Info("delete object success", "key", key)
								return
							}
							slog.Warn("delete object retry", "attempt", i+1, "key", key, "error", err)
							deleteErr = err
							time.Sleep(3 * time.Second)
						}

						slog.Error("delete object failed", "key", key, "error", deleteErr)
						select {
						case errChan <- fmt.Errorf("delete object error, key: %s: %w", key, deleteErr):
						default:
						}
						cancel()
					}(currentKey)
				}

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

func (s *ProblemService) UpdateProblemDescription(
	ctx context.Context,
	id int,
	description string,
) error {
	return foundationdao.GetProblemDao().UpdateProblemDescription(ctx, id, description)
}

func (s *ProblemService) UpdateProblemJudgeInfo(
	ctx context.Context,
	id int,
	judgeType foundationjudge.JudgeType,
	md5 string,
) error {
	return foundationdao.GetProblemDao().UpdateProblemJudgeInfo(ctx, id, judgeType, md5)
}
