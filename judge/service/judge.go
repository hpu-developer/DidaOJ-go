package service

import (
	"meta/cron"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
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
