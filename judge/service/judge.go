package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	"io"
	"judge/config"
	"log/slog"
	"meta/cron"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	metapath "meta/meta-path"
	metastring "meta/meta-string"
	"meta/routine"
	"meta/singleton"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// 需要保证只有一个goroutine在处理判题数据
type judgeDataDownloadEntry struct {
	mu  sync.Mutex
	ref int32
}

type JudgeService struct {
	runningTasks      atomic.Int32
	judgeDataDownload sync.Map
}

var singletonJudgeService = singleton.Singleton[JudgeService]{}

func GetJudgeService() *JudgeService {
	return singletonJudgeService.GetInstance(
		func() *JudgeService {
			return &JudgeService{}
		},
	)
}

func (s *JudgeService) Start() error {
	c := cron.NewWithSeconds()
	_, err := c.AddFunc(
		"* * * * * ?", func() {
			// 每秒运行一次任务
			err := s.handleStart()
			if err != nil {
				metapanic.ProcessError(err)
				return
			}
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "error adding function to cron")
	}

	c.Start()

	return nil
}

func (s *JudgeService) handleStart() error {
	maxJob := config.GetConfig().MaxJob
	if int(s.runningTasks.Load()) >= maxJob {
		return nil
	}
	ctx := context.Background()
	jobs, err := foundationdao.GetJudgeJobDao().GetJudgeJobListPendingJudge(ctx, maxJob)
	if err != nil {
		return metaerror.Wrap(err, "failed to get judge job list")
	}
	s.runningTasks.Add(int32(len(jobs)))
	for _, job := range jobs {
		routine.SafeGo(fmt.Sprintf("RunningJudgeJob_%d", job.Id), func() error {
			defer s.runningTasks.Add(-1)
			err = s.startJudgeTask(job)
			if err != nil {
				err := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(ctx, job.Id, foundationjudge.JudgeStatusJudgeFail)
				if err != nil {
					metapanic.ProcessError(err)
				}
				return err
			}
			return nil
		})
	}
	return nil
}

func (s *JudgeService) startJudgeTask(job *foundationmodel.JudgeJob) error {
	ctx := context.Background()

	err := foundationdao.GetJudgeJobDao().StartProcessJudgeJob(ctx, job.Id, config.GetConfig().Judger)
	if err != nil {
		return metaerror.Wrap(err, "failed to start process judge job")
	}
	problem, err := foundationdao.GetProblemDao().GetProblem(ctx, job.ProblemId)
	if err != nil {
		return metaerror.Wrap(err, "failed to get problem")
	}
	err = s.updateJudgeData(ctx, problem.Id, problem.JudgeMd5)
	if err != nil {
		return metaerror.Wrap(err, "failed to update judge data")
	}
	execFileId, extraMessage, err := s.compileCode(job)
	if extraMessage != "" {
		markErr := foundationdao.GetJudgeJobDao().MarkJudgeJobCompileMessage(ctx, job.Id, extraMessage)
		if markErr != nil {
			metapanic.ProcessError(markErr)
		}
	}
	if err != nil {
		return err
	}
	if execFileId == nil {
		err := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(ctx, job.Id, foundationjudge.JudgeStatusCE)
		if err != nil {
			metapanic.ProcessError(err)
		}
		return nil
	}
	slog.Info("compile code success", "job", job.Id, "execFileId", *execFileId)
	err = foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeStatus(ctx, job.Id, foundationjudge.JudgeStatusRunning)
	if err != nil {
		metapanic.ProcessError(err)
	}
	err = s.runJudgeTask(ctx, job, problem.TimeLimit, problem.MemoryLimit, *execFileId)

	//	runUrl := metahttp.UrlJoin(config.GetConfig().GoJudgeUrl, "run")
	//
	//	// 准备请求数据
	//	data := map[string]interface{}{
	//		"cmd": []map[string]interface{}{
	//			{
	//				"args": []string{"/usr/bin/g++", "a.cc", "-o", "a"},
	//				"env":  []string{"PATH=/usr/bin:/bin"},
	//				"files": []map[string]interface{}{
	//					{"content": ""},
	//					{"name": "stdout", "max": 10240},
	//					{"name": "stderr", "max": 10240},
	//				},
	//				"cpuLimit":    10000000000,
	//				"memoryLimit": 104857600,
	//				"procLimit":   50,
	//				"copyIn": map[string]interface{}{
	//					"a.cc": map[string]interface{}{
	//						"content": `#include <iostream>
	//using namespace std;
	//int main() {
	//    int a, b;
	//    cin >> a >> b;
	//    cout << a + b << endl;
	//}`,
	//					},
	//				},
	//				"copyOut":       []string{"stdout", "stderr"},
	//				"copyOutCached": []string{"a"},
	//			},
	//		},
	//	}
	//
	//	// 编码成 JSON
	//	jsonData, err := json.Marshal(data)
	//	if err != nil {
	//		return err
	//	}
	//
	//	// 发起 POST 请求
	//	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	//	if err != nil {
	//		return err
	//	}
	//	defer resp.Body.Close()
	//
	//	// 可根据需要处理返回，比如判断状态码
	//	if resp.StatusCode != http.StatusOK {
	//		return metaerror.New("unexpected status code: %d", resp.StatusCode)
	//	}
	//
	//	responseString := new(bytes.Buffer)
	//	_, err = responseString.ReadFrom(resp.Body)
	//	if err != nil {
	//		return metaerror.Wrap(err)
	//	}
	//
	//	slog.Info(responseString.String())
	return nil
}

func (s *JudgeService) updateJudgeData(ctx context.Context, problemId string, md5 string) error {
	val, _ := s.judgeDataDownload.LoadOrStore(problemId, &judgeDataDownloadEntry{})
	e := val.(*judgeDataDownloadEntry)
	atomic.AddInt32(&e.ref, 1)
	defer func() {
		if atomic.AddInt32(&e.ref, -1) == 0 {
			s.judgeDataDownload.Delete(problemId)
		}
	}()
	e.mu.Lock()
	defer e.mu.Unlock()
	judgeMd5FilePath := path.Join(".judge_data", problemId, "md5.txt")
	content, err := metastring.GetStringFromOpenFile(judgeMd5FilePath)
	if err != nil || strings.TrimSpace(content) != strings.TrimSpace(md5) {
		return s.downloadJudgeData(ctx, problemId)
	}
	return nil
}

func (s *JudgeService) downloadJudgeData(ctx context.Context, problemId string) error {

	slog.Info("downloading judge data", "problemId", problemId)

	// 删除旧的判题数据
	judgeDataDir := path.Join(".judge_data", problemId)
	err := os.RemoveAll(judgeDataDir)

	// 初始化 R2 连接（这里用 AWS SDK）
	sess, err := session.NewSession(&aws.Config{
		Region:           aws.String("auto"),                           // R2一般写 auto
		Endpoint:         aws.String(config.GetConfig().JudgeData.Url), // 替换成你的 R2 Endpoint
		S3ForcePathStyle: aws.Bool(true),                               // R2要求这个必须 true
		Credentials: credentials.NewStaticCredentials(config.GetConfig().JudgeData.Key,
			config.GetConfig().JudgeData.Secret,
			config.GetConfig().JudgeData.Token),
	})
	if err != nil {
		return metaerror.Wrap(err, "failed to create session")
	}

	s3Client := s3.New(sess)

	// 1. 列出 problemId 目录下的所有对象
	input := &s3.ListObjectsV2Input{
		Bucket: aws.String("didaoj-judge"),
		Prefix: aws.String(problemId + "/"), // 确保带 `/`，只列出这个目录下的
	}

	var wg sync.WaitGroup
	var mu sync.Mutex
	var downloadErr error

	err = s3Client.ListObjectsV2PagesWithContext(ctx, input, func(page *s3.ListObjectsV2Output, lastPage bool) bool {
		for _, obj := range page.Contents {
			wg.Add(1)
			go func(obj *s3.Object) {
				defer wg.Done()
				localPath := path.Join(".judge_data", *obj.Key)
				err := s.downloadObject(ctx, s3Client, "didaoj-judge", *obj.Key, localPath)
				if err != nil {
					mu.Lock()
					if downloadErr == nil {
						downloadErr = err
					}
					mu.Unlock()
				}
			}(obj)
		}
		return true
	})
	if err != nil {
		return metaerror.Wrap(err, "failed to list objects")
	}

	// 等待所有下载完成
	wg.Wait()

	// 如果有任何错误，返回
	if downloadErr != nil {
		return downloadErr
	}

	return nil
}

// 单独抽一个下载单个对象的方法
func (s *JudgeService) downloadObject(ctx context.Context, s3Client *s3.S3, bucket, key string, localPath string) error {
	getObjInput := &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}
	result, err := s3Client.GetObjectWithContext(ctx, getObjInput)
	if err != nil {
		return fmt.Errorf("failed to get object %s: %w", key, err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(result.Body)
	err = os.MkdirAll(filepath.Dir(localPath), 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory for %s: %w", localPath, err)
	}
	outFile, err := os.Create(localPath)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", localPath, err)
	}
	defer func(outFile *os.File) {
		err := outFile.Close()
		if err != nil {
			metapanic.ProcessError(metaerror.Wrap(err))
		}
	}(outFile)
	_, err = io.Copy(outFile, result.Body)
	if err != nil {
		return metaerror.Wrap(err, "failed to save object %s", key)
	}
	return nil
}

func (s *JudgeService) compileCode(job *foundationmodel.JudgeJob) (*string, string, error) {
	slog.Info("compile code", "job", job.Id)
	runUrl := metahttp.UrlJoin(config.GetConfig().GoJudgeUrl, "run")
	// 准备请求数据
	data := map[string]interface{}{
		"cmd": []map[string]interface{}{
			{
				"args": []string{"/usr/bin/g++", "a.cc", "-o", "a"},
				"env":  []string{"PATH=/usr/bin:/bin"},
				"files": []map[string]interface{}{
					{"content": ""},
					{"name": "stdout", "max": 10240},
					{"name": "stderr", "max": 10240},
				},
				"cpuLimit":    10000000000,
				"memoryLimit": 104857600,
				"procLimit":   50,
				"copyIn": map[string]interface{}{
					"a.cc": map[string]interface{}{
						"content": job.Code,
					},
				},
				"copyOut":       []string{"stdout", "stderr"},
				"copyOutCached": []string{"a"},
			},
		},
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return nil, "compile failed, system error.", err
	}
	resp, err := http.Post(runUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, "compile failed, upload file error.", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			metapanic.ProcessError(err)
		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, "compile failed, upload file response error.", metaerror.New("unexpected status code: %d", resp.StatusCode)
	}
	var responseDataList []struct {
		Status string `json:"status"`
		Files  struct {
			Stderr string `json:"stderr"`
			Stdout string `json:"stdout"`
		} `json:"files"`
		FileIds struct {
			A string `json:"a"`
		}
	}
	err = json.NewDecoder(resp.Body).Decode(&responseDataList)
	if err != nil {
		return nil, "compile failed, upload file response parse error.", metaerror.Wrap(err, "failed to decode response")
	}
	if len(responseDataList) != 1 {
		return nil, "compile failed, compile response data error.", metaerror.New("unexpected response length: %d", len(responseDataList))
	}
	responseData := responseDataList[0]
	errorMessage := responseData.Files.Stderr + "\n" + responseData.Files.Stdout
	if responseData.Status != "Accepted" {
		return nil, errorMessage, nil
	}
	return &responseData.FileIds.A, errorMessage, nil
}

func (s *JudgeService) runJudgeTask(ctx context.Context, job *foundationmodel.JudgeJob, timeLimit int, memoryLimit int, execFileId string) error {
	problemId := job.ProblemId
	val, _ := s.judgeDataDownload.LoadOrStore(problemId, &judgeDataDownloadEntry{})
	e := val.(*judgeDataDownloadEntry)
	atomic.AddInt32(&e.ref, 1)
	defer func() {
		if atomic.AddInt32(&e.ref, -1) == 0 {
			s.judgeDataDownload.Delete(problemId)
		}
	}()
	e.mu.Lock()
	defer e.mu.Unlock()

	judgeDataDir := path.Join(".judge_data", problemId)
	files, err := os.ReadDir(judgeDataDir)
	if err != nil {
		return metaerror.Wrap(err, "failed to read judge data dir")
	}
	// TODO获取rule.yaml文件

	enableRule := false

	hasInFiles := make(map[string]bool)
	var Files []string

	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".in") {
			hasInFiles[metapath.GetBaseName(file.Name())] = true
		} else if strings.HasSuffix(file.Name(), ".out") {
			Files = append(Files, metapath.GetBaseName(file.Name()))
		}
	}
	taskCount := 0
	sort.Slice(Files, func(i, j int) bool {
		return Files[i] < Files[j]
	})
	for _, file := range Files {
		if !hasInFiles[file] {
			continue
		}
		taskCount++
	}
	err = foundationdao.GetJudgeJobDao().MarkJudgeJobTaskTotal(ctx, job.Id, taskCount)
	if err != nil {
		metapanic.ProcessError(err)
	}

	acTask := 0

	for _, file := range Files {
		runUrl := metahttp.UrlJoin(config.GetConfig().GoJudgeUrl, "run")
		inContent, err := metastring.GetStringFromOpenFile(path.Join(judgeDataDir, file+".in"))
		if err != nil {
			return err
		}
		data := map[string]interface{}{
			"cmd": []map[string]interface{}{
				{
					"args": []string{"a"},
					"env":  []string{"PATH=/usr/bin:/bin"},
					"files": []map[string]interface{}{
						{"content": inContent},
						{"name": "stdout", "max": 10240},
						{"name": "stderr", "max": 10240},
					},
					"cpuLimit":    timeLimit * 1000000,
					"memoryLimit": memoryLimit * 1024,
					"procLimit":   50,
					"copyIn": map[string]interface{}{
						"a": map[string]interface{}{
							"fileId": execFileId,
						},
					},
				},
			},
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return metaerror.Wrap(err)
		}
		resp, err := http.Post(runUrl, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return metaerror.Wrap(err)
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				metapanic.ProcessError(err)
			}
		}(resp.Body)
		if resp.StatusCode != http.StatusOK {
			return metaerror.New("unexpected status code: %d", resp.StatusCode)
		}
		var responseDataList []struct {
			Status string `json:"status"`
			Files  struct {
				Stderr string `json:"stderr"`
				Stdout string `json:"stdout"`
			} `json:"files"`
		}
		err = json.NewDecoder(resp.Body).Decode(&responseDataList)
		if err != nil {
			return metaerror.Wrap(err, "failed to decode response")
		}
		if len(responseDataList) != 1 {
			return metaerror.New("unexpected response length: %d", len(responseDataList))
		}
		responseData := responseDataList[0]
		if responseData.Status != "Accepted" {
			continue
		}

		outContent, err := metastring.GetStringFromOpenFile(path.Join(judgeDataDir, file+".out"))
		if err != nil {
			return err
		}

		err = foundationdao.GetJudgeJobDao().AddJudgeJobTaskCurrent(ctx, job.Id)
		if err != nil {
			return err
		}
	}

	return nil
}
