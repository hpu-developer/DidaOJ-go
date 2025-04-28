package service

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationmodel "foundation/foundation-model"
	"io"
	"judge/config"
	"meta/cron"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	metastring "meta/meta-string"
	"meta/routine"
	"meta/singleton"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

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
				err := foundationdao.GetJudgeJobDao().MarkJudgeJobJudgeFailed(ctx, job.Id)
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

	err := foundationdao.GetJudgeJobDao().StartProcessJudgeJob(ctx, job.Id)
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

	err = s.compileCode(job)
	if err != nil {
		return err
	}

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
	// 需要保证只有一个goroutine在处理判题数据
	type judgeDataDownloadEntry struct {
		mu  sync.Mutex
		ref int32
	}
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

func (s *JudgeService) compileCode(job *foundationmodel.JudgeJob) error {

	//	{
	//		"cmd": [{
	//	"args": ["/usr/bin/g++", "a.cc", "-o", "a"],
	//	"env": ["PATH=/usr/bin:/bin"],
	//	"files": [{
	//	"content": ""
	//	}, {
	//	"name": "stdout",
	//	"max": 10240
	//	}, {
	//	"name": "stderr",
	//	"max": 10240
	//	}],
	//	"cpuLimit": 10000000000,
	//	"memoryLimit": 104857600,
	//	"procLimit": 50,
	//	"copyIn": {
	//	"a.cc": {
	//	"content": "#include <iostream>\nusing namespace std;\nint main() {\nint a, b;\ncin >> a >> b;\ncout << a + b << endl;\n}"
	//	}
	//	},
	//	"copyOut": ["stdout", "stderr"],
	//	"copyOutCached": ["a"]
	//	}]
	//}
	runUrl := metahttp.UrlJoin(config.GetConfig().GoJudgeUrl, "run")
	// 准备请求数据
	data := map[string]interface{}{
		"cmd": []map[string]interface{}{
			{
				"args": []string{"/usr/bin/g++", "a.cc", "-o", "a"},
				"env":  []string{"PATH=/usr/bin:/bin"},
				"files": []map[string]interface{}{
					{"content": job.Code},
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
		return err
	}
	resp, err := http.Post(runUrl, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
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

	return nil
}
