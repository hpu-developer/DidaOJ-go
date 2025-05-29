package controller

import (
	"bytes"
	"fmt"
	foundationerrorcode "foundation/error-code"
	foundationauth "foundation/foundation-auth"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationoj "foundation/foundation-oj"
	foundationservice "foundation/foundation-service"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"gopkg.in/yaml.v3"
	"log/slog"
	cfr2 "meta/cf-r2"
	metacontroller "meta/controller"
	"meta/error-code"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metamath "meta/meta-math"
	metamd5 "meta/meta-md5"
	metapanic "meta/meta-panic"
	metapath "meta/meta-path"
	"meta/meta-response"
	metaslice "meta/meta-slice"
	metastring "meta/meta-string"
	metatime "meta/meta-time"
	metazip "meta/meta-zip"
	"meta/set"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"web/config"
	weberrorcode "web/error-code"
	"web/request"
	"web/service"
)

type ProblemJudgeData struct {
	Key          string     `json:"key"`
	Size         *int64     `json:"size"`
	LastModified *time.Time `json:"last_modified"`
}

type ProblemController struct {
	metacontroller.Controller
}

func (c *ProblemController) Get(ctx *gin.Context) {
	problemService := foundationservice.GetProblemService()
	id := ctx.Query("id")
	isContest := false
	if id == "" {
		contestIdStr := ctx.Query("contest_id")
		if contestIdStr == "" {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		contestId, err := strconv.Atoi(contestIdStr)
		if err != nil || contestId <= 0 {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		problemIndexStr := ctx.Query("problem_index")
		if problemIndexStr == "" {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		problemIndex, err := strconv.Atoi(problemIndexStr)
		if err != nil || problemIndex <= 0 {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		idPtr, err := problemService.GetProblemIdByContest(ctx, contestId, problemIndex)
		if err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		if idPtr == nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
		id = *idPtr
		isContest = true
	}

	userId, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)

	problem, err := problemService.GetProblemView(ctx, id, userId, ok)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	var tags []*foundationmodel.ProblemTag
	if isContest {
		// 比赛时隐藏一些信息
		problem.Id = ""
		problem.Source = ""
		problem.OriginOj = nil
		problem.OriginId = nil
		problem.OriginUrl = nil
	} else {
		if problem.Tags != nil {
			tags, err = problemService.GetProblemTagByIds(ctx, problem.Tags)
			if err != nil {
				metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
				return
			}
		}
	}
	responseData := struct {
		Problem *foundationmodel.Problem      `json:"problem"`
		Tags    []*foundationmodel.ProblemTag `json:"tags,omitempty"`
	}{
		Problem: problem,
		Tags:    tags,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetList(ctx *gin.Context) {
	pageStr := ctx.DefaultQuery("page", "1")
	pageSizeStr := ctx.DefaultQuery("page_size", "50")
	page, err := strconv.Atoi(pageStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	pageSize, err := strconv.Atoi(pageSizeStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if pageSize != 50 && pageSize != 100 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	problemService := foundationservice.GetProblemService()
	oj := ctx.Query("oj")
	if oj != "" {
		oj = foundationoj.GetOriginOjKey(oj)
		if oj == "" {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
	}
	title := ctx.Query("title")
	tag := ctx.Query("tag")
	var list []*foundationmodel.Problem
	var totalCount int
	var problemStatus map[string]foundationmodel.ProblemAttemptStatus
	userId, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if userId > 0 {
		private := ctx.Query("private") != "0"
		list, totalCount, problemStatus, err = problemService.GetProblemListWithUser(ctx, userId, ok, oj, title, tag, private, page, pageSize)
	} else {
		list, totalCount, err = problemService.GetProblemList(ctx, oj, title, tag, page, pageSize)
	}
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	responseData := struct {
		Time                 time.Time                                       `json:"time"`
		TotalCount           int                                             `json:"total_count"`
		List                 []*foundationmodel.Problem                      `json:"list"`
		ProblemAttemptStatus map[string]foundationmodel.ProblemAttemptStatus `json:"problem_attempt_status,omitempty"`
	}{
		Time:                 metatime.GetTimeNow(),
		TotalCount:           totalCount,
		List:                 list,
		ProblemAttemptStatus: problemStatus,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetRecommend(ctx *gin.Context) {
	userId, hasAuth, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if userId <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problemService := foundationservice.GetProblemService()
	problemId := ctx.Query("problem_id")
	list, err := problemService.GetProblemRecommend(ctx, userId, hasAuth, problemId)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	var tags []*foundationmodel.ProblemTag
	if len(list) > 0 {
		tagIdSet := set.New[int]()
		for _, problem := range list {
			if problem.Tags != nil {
				for _, tagId := range problem.Tags {
					tagIdSet.Add(tagId)
				}
			}
		}
		var tagIds []int
		tagIdSet.Foreach(func(tagId *int) bool {
			tagIds = append(tagIds, *tagId)
			return true
		})
		tags, err = foundationservice.GetProblemService().GetProblemTagByIds(ctx, tagIds)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
	}
	responseData := struct {
		List []*foundationmodel.Problem    `json:"list"`
		Tags []*foundationmodel.ProblemTag `json:"tags,omitempty"`
	}{
		List: list,
		Tags: tags,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetTagList(ctx *gin.Context) {
	problemService := foundationservice.GetProblemService()
	maxCountStr := ctx.DefaultQuery("max_count", "-1")
	maxCount, err := strconv.Atoi(maxCountStr)
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	list, totalCount, err := problemService.GetProblemTagList(ctx, maxCount)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	responseData := struct {
		Time       time.Time                     `json:"time"`
		TotalCount int                           `json:"total_count"`
		List       []*foundationmodel.ProblemTag `json:"list"`
	}{
		Time:       metatime.GetTimeNow(),
		TotalCount: totalCount,
		List:       list,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}

func (c *ProblemController) GetJudge(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problemService := foundationservice.GetProblemService()
	problem, err := problemService.GetProblemJudge(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	r2Client := cfr2.GetSubsystem().GetClient("judge-data")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	problemId := problem.Id
	prefixKey := filepath.ToSlash(problemId + "/")
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("didaoj-judge"),
		Prefix: aws.String(prefixKey),
	}

	var judges []*ProblemJudgeData

	err = r2Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			judgeData := &ProblemJudgeData{
				Key:          strings.TrimPrefix(*obj.Key, prefixKey),
				Size:         obj.Size,
				LastModified: obj.LastModified,
			}
			judges = append(judges, judgeData)
		}
		return true
	})
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	responseData := struct {
		Problem *foundationmodel.Problem `json:"problem"`
		Judges  []*ProblemJudgeData      `json:"judges"`
	}{
		Problem: problem,
		Judges:  judges,
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, responseData)
}
func (c *ProblemController) GetJudgeDataDownload(ctx *gin.Context) {
	id := ctx.Query("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	key := ctx.Query("key")
	if key == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	// 鉴权
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	// 获取题目信息
	problemService := foundationservice.GetProblemService()
	problem, err := problemService.GetProblemJudge(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	// 获取 R2 客户端
	r2Client := cfr2.GetSubsystem().GetClient("judge-data")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	// 生成预签名链接
	objectKey := filepath.ToSlash(path.Join(id, key))
	req, _ := r2Client.GetObjectRequest(&s3.GetObjectInput{
		Bucket: aws.String("didaoj-judge"),
		Key:    aws.String(objectKey),
	})
	expire := 10 * time.Minute
	urlStr, err := req.Presign(expire)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, urlStr)
}

func (c *ProblemController) PostCrawl(ctx *gin.Context) {
	var requestData struct {
		OJ string `json:"oj",binding:"required"`
		Id string `json:"id",binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.OJ == "" || requestData.Id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	oj := strings.ToLower(requestData.OJ)
	id := strings.TrimSpace(requestData.Id)
	if oj == "didaoj" {
		ok, err := foundationservice.GetProblemService().HasProblem(ctx, id)
		if err != nil {
			return
		}
		if !ok {
			metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
			return
		}
		metaresponse.NewResponse(ctx, metaerrorcode.Success, id)
		return
	}
	newId, err := service.GetProblemCrawlService().PostCrawlProblem(ctx, oj, id)
	if err != nil {
		metaresponse.NewResponseError(ctx, err)
		return
	}
	if newId == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, newId)
}

func (c *ProblemController) PostJudgeData(ctx *gin.Context) {
	id := ctx.PostForm("id")
	if id == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	_, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	problemService := foundationservice.GetProblemService()
	problem, err := problemService.GetProblemJudge(ctx, id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if problem == nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}
	file, err := ctx.FormFile("zip")
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError)
		return
	}
	tempDir, err := os.MkdirTemp("", "didaoj-judge-data-*")
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}
	defer func(path string) {
		err := os.RemoveAll(path)
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err, "<UNK>: "+path))
		}
	}(tempDir)
	uploadedPath := filepath.Join(tempDir, file.Filename)
	if err := ctx.SaveUploadedFile(file, uploadedPath); err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError)
		return
	}
	unzipDir := filepath.Join(tempDir, "unzipped")
	if err := metazip.UzipFile(uploadedPath, unzipDir); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	// 如果包含文件夹，认为失败
	err = filepath.Walk(unzipDir, func(path string, info os.FileInfo, err error) error {
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
	})
	if err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError)
		return
	}

	judgeType := foundationjudge.JudgeTypeNormal

	var jobConfig foundationjudge.JudgeJobConfig

	// 解析rule.yaml
	ruleFile := filepath.Join(unzipDir, "rule.yaml")
	yamlFile, err := os.ReadFile(ruleFile)
	if err == nil {
		err = yaml.Unmarshal(yamlFile, &jobConfig)
		if err != nil {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
		}
	} else {
		if !os.IsNotExist(err) {
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
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
		goJudgeUrl := config.GetConfig().GoJudge.Url
		runUrl := metahttp.UrlJoin(goJudgeUrl, "run")

		language := foundationjudge.GetLanguageByKey(jobConfig.SpecialJudge.Language)
		if !foundationjudge.IsValidJudgeLanguage(int(language)) {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}

		// 考虑编译机性能影响，暂时仅允许部分语言
		if !foundationjudge.IsValidSpecialJudgeLanguage(language) {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}

		codeFilePath := filepath.Join(unzipDir, jobConfig.SpecialJudge.Source)
		codeContent, err := metastring.GetStringFromOpenFile(codeFilePath)
		if err != nil {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}

		jobKey := uuid.New().String()

		execFileIds, extraMessage, compileStatus, err := foundationjudge.CompileCode(jobKey, runUrl, language, codeContent, nil)
		if extraMessage != "" {
			slog.Warn("judge compile", "extraMessage", extraMessage, "compileStatus", compileStatus)
		}
		if compileStatus != foundationjudge.JudgeStatusAC {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}
		if err != nil {
			metapanic.ProcessError(err)
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
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
			metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
			return
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
		taskKeys = metaslice.RemoveAllFunc(taskKeys, func(key string) bool {
			return !hasInFiles[key] && !hasOutFiles[key]
		})
		taskCount := len(taskKeys)

		if taskCount <= 0 {
			metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
			return
		}

		sort.Slice(taskKeys, func(i, j int) bool {
			return taskKeys[i] < taskKeys[j]
		})

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
					metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
					return
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

	totalScore := 0
	for _, taskConfig := range jobConfig.Tasks {
		totalScore += taskConfig.Score
	}
	if totalScore <= 0 {
		totalScore = 100
		averageScore := totalScore / taskCount
		for i, taskConfig := range jobConfig.Tasks {
			if i != taskCount-1 {
				taskConfig.Score = averageScore
			} else {
				taskConfig.Score = totalScore - averageScore*(taskCount-1)
			}
		}
	} else {
		//把totalScore转为0~100
		rate := 100.0 / float64(totalScore)
		totalScore = 100
		sumScore := 0
		for i, taskConfig := range jobConfig.Tasks {
			if i != taskCount-1 {
				taskConfig.Score = int(float64(taskConfig.Score) * rate)
				sumScore += taskConfig.Score
			} else {
				taskConfig.Score = totalScore - sumScore
			}
		}
	}

	// 重新生成一个rule.yaml
	ruleFile = filepath.Join(unzipDir, "rule.yaml")
	yamlData, err := yaml.Marshal(jobConfig)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	err = os.WriteFile(ruleFile, yamlData, 0644)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	// 把所有文件的换行改为Linux格式
	err = filepath.Walk(unzipDir, func(path string, info os.FileInfo, err error) error {
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
	})
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	var files []string
	err = filepath.Walk(unzipDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		files = append(files, path)
		return nil
	})

	judgeDataMd5, err := metamd5.MultiFileMD5(files)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	slog.Info("judge data md5", "md5", judgeDataMd5)

	if problem.JudgeMd5 != nil && *problem.JudgeMd5 == judgeDataMd5 {
		metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
		return
	}

	// 上传r2
	r2Client := cfr2.GetSubsystem().GetClient("judge-data")
	if r2Client == nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	// 遍历解压目录并上传文件
	err = filepath.Walk(unzipDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		relativePath, err := filepath.Rel(unzipDir, path)
		if err != nil {
			return err
		}
		key := filepath.ToSlash(filepath.Join(id, judgeDataMd5, relativePath))
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func(file *os.File) {
			err := file.Close()
			if err != nil {
				metapanic.ProcessError(metaerror.Wrap(err, "close file error"))
			}
		}(file)
		slog.Info("put object start", "key", key)
		_, err = r2Client.PutObjectWithContext(ctx, &s3.PutObjectInput{
			Bucket: aws.String("didaoj-judge"),
			Key:    aws.String(key),
			Body:   file,
		})
		if err != nil {
			slog.Info("put object error", "key", key)
			return metaerror.Wrap(err, "put object error, key:%s", key)
		}
		slog.Info("put object success", "key", key)
		return nil
	})
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, fmt.Sprintf("上传失败: %v", err))
		return
	}
	// 删除旧的路径
	if problem.JudgeMd5 != nil {
		prefix := filepath.ToSlash(path.Join(id, *problem.JudgeMd5))
		input := &s3.ListObjectsV2Input{
			Bucket: aws.String("didaoj-judge"),
			Prefix: aws.String(prefix),
		}
		err = r2Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
			for _, obj := range page.Contents {
				_, err := r2Client.DeleteObjectWithContext(ctx, &s3.DeleteObjectInput{
					Bucket: aws.String("didaoj-judge"),
					Key:    obj.Key,
				})
				if err != nil {
					metapanic.ProcessError(metaerror.Wrap(err, "delete object error, key:%s", obj.Key))
					return false
				}
			}
			return true
		})
	}
	err = problemService.UpdateProblemJudgeInfo(ctx, id, judgeType, judgeDataMd5)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	metaresponse.NewResponse(ctx, metaerrorcode.Success, nil)
}

func (c *ProblemController) PostCreate(ctx *gin.Context) {
	var requestData request.ProblemEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Title == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.TimeLimit <= 0 || requestData.MemoryLimit <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	userId, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	ok, err = foundationservice.GetProblemService().HasProblemTitle(ctx, requestData.Title)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if ok {
		metaresponse.NewResponse(ctx, weberrorcode.ProblemTitleDuplicate, nil)
		return
	}
	problemId, err := foundationservice.GetProblemService().PostCreate(ctx, userId, &requestData)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, problemId)
}

func (c *ProblemController) PostEdit(ctx *gin.Context) {
	var requestData request.ProblemEdit
	if err := ctx.ShouldBindJSON(&requestData); err != nil {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.Title == "" {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}
	if requestData.TimeLimit <= 0 || requestData.MemoryLimit <= 0 {
		metaresponse.NewResponse(ctx, foundationerrorcode.ParamError, nil)
		return
	}

	userId, ok, err := foundationservice.GetUserService().CheckUserAuth(ctx, foundationauth.AuthTypeManageProblem)
	if err != nil {
		metapanic.ProcessError(err)
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}
	if !ok {
		metaresponse.NewResponse(ctx, foundationerrorcode.AuthError, nil)
		return
	}

	hasProblem, err := foundationservice.GetProblemService().HasProblem(ctx, requestData.Id)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}
	if !hasProblem {
		metaresponse.NewResponse(ctx, foundationerrorcode.NotFound, nil)
		return
	}

	updateTime, err := foundationservice.GetProblemService().PostEdit(ctx, userId, &requestData)
	if err != nil {
		metaresponse.NewResponse(ctx, metaerrorcode.CommonError, nil)
		return
	}

	metaresponse.NewResponse(ctx, metaerrorcode.Success, updateTime)
}
