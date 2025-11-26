package service

import (
	"context"
	"errors"
	"fmt"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	"io"
	"judge/config"
	gojudge "judge/go-judge"
	"log"
	"log/slog"
	"meta/cron"
	metaerror "meta/meta-error"
	metahttp "meta/meta-http"
	metapanic "meta/meta-panic"
	"meta/metaroutine"
	"meta/singleton"
	"net/http"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"
)

// 需要保证只有一个goroutine在处理判题数据
type botMutexEntry struct {
	mu  sync.Mutex
	ref int32
}

type BotService struct {
	requestMutex    sync.Mutex
	botRunningTasks atomic.Int32

	// 防止因重判等情况多次获取到了同一个判题任务（不过多个判题机则靠key来忽略）
	botJobMutexMap sync.Map

	judgeFileIds map[int]string

	goJudgeClient *http.Client
}

var singletonBotService = singleton.Singleton[BotService]{}

func GetBotService() *BotService {
	return singletonBotService.GetInstance(
		func() *BotService {
			s := &BotService{}
			s.goJudgeClient = &http.Client{
				Transport: &http.Transport{
					MaxIdleConns:        100,
					MaxIdleConnsPerHost: 100,
					MaxConnsPerHost:     100,
					IdleConnTimeout:     90 * time.Second,
				},
				Timeout: 60 * time.Second, // 请求整体超时
			}
			return s
		},
	)
}

func (s *BotService) Start() error {

	c := cron.NewWithSeconds()
	_, err := c.AddFunc(
		"* * * * * ?", func() {
			// 每秒检查一次任务是否能运行
			s.checkStartBot()
		},
	)
	if err != nil {
		return metaerror.Wrap(err, "error adding function to cron")
	}

	c.Start()

	return nil
}

func (s *BotService) checkStartBot() {
	err := s.handleStart()
	if err != nil {
		metapanic.ProcessError(err)
	}
}

func (s *BotService) handleStart() error {

	// 如果没开启评测，停止判题
	if !GetStatusService().IsEnableJudge() {
		return nil
	}
	// 如果上报状态报错，停止判题
	if GetStatusService().IsReportError() {
		return nil
	}

	// 保证同时只有一个handleStart
	if !s.requestMutex.TryLock() {
		return nil
	}
	defer s.requestMutex.Unlock()

	// 连接url并发送测试请求
	// 五子棋评测程序将在其他地方调用
	// 运行五子棋评测逻辑
	err := s.testBase()
	if err != nil {
		return metaerror.Wrap(err, "failed to test base")
	}

	return nil
}

func (s *BotService) testBase() error {
	judgeFileId, err := s.compileJudge(1)
	if err != nil {
		return metaerror.Wrap(err, "failed to compile special judge")
	}
	if judgeFileId == "" {
		return metaerror.New("special judge compile failed")
	}
	s.runGame(judgeFileId)

	return nil
}

func (s *BotService) getJudgeFileId(gameId int) string {
	if s.judgeFileIds == nil {
		return ""
	}
	return s.judgeFileIds[gameId]
}

func (s *BotService) compileJudge(gameId int) (string, error) {

	specialFileId := s.getJudgeFileId(gameId)
	if specialFileId != "" {
		return specialFileId, nil
	}

	ctx := context.Background()
	code, err := foundationdao.GetBotGameDao().GetJudgeCode(ctx, gameId)
	if err != nil {
		return "", metaerror.Wrap(err, "failed to start process judge job")
	}

	runUrl := metahttp.UrlJoin(config.GetConfig().GoJudge.Url, "run")

	execFileIds, extraMessage, compileStatus, err := foundationjudge.CompileCode(
		s.goJudgeClient,
		"bot_1_judge",
		runUrl,
		foundationjudge.JudgeLanguageGolang,
		code,
		GetJudgeService().configFileIds,
		false,
		true,
	)
	if extraMessage != "" {
		slog.Warn("judge compile", "extraMessage", extraMessage, "compileStatus", compileStatus)
	}
	if compileStatus != foundationjudge.JudgeStatusAC {
		return "", metaerror.New("compile special judge failed: %s", extraMessage)
	}
	if err != nil {
		return "", metaerror.Wrap(err, "failed to compile special judge")
	}
	var ok bool
	specialFileId, ok = execFileIds["a"]
	if !ok {
		return "", metaerror.New("special judge compile failed, fileId not found")
	}
	if s.judgeFileIds == nil {
		s.judgeFileIds = make(map[int]string)
	}
	s.judgeFileIds[gameId] = specialFileId
	return specialFileId, nil
}

func (s *BotService) runGame(judgeFileId string) error {
	wsURL := "http://127.0.0.1:30000/stream"
	streamClient := s.newWebsocket([]string{}, wsURL)
	if streamClient == nil {
		return metaerror.New("创建WebSocket连接失败")
	}
	defer streamClient.Close()

	var args []string
	var copyIns map[string]gojudge.CmdFile
	args = []string{"a"}
	copyIns = map[string]gojudge.CmdFile{
		"a": {
			FileID: judgeFileId,
		},
	}

	cpuLimit := uint64(5000000000)
	memoryLimit := uint64(104857600)

	req := &gojudge.RunRequest{
		RequestID: fmt.Sprintf("%d-%s-%d", 1, time.Now().Format("20060102150405"), time.Now().UnixNano()),
		Cmd: []gojudge.Cmd{
			{
				Args: args,
				Env:  []string{"PATH=/usr/bin:/bin"},
				Files: []*gojudge.CmdFile{
					{StreamIn: true},
					{StreamOut: true},
				},
				CPULimit:    cpuLimit,    // 5秒
				MemoryLimit: memoryLimit, // 100MB
				ProcLimit:   50,
				CopyIn:      copyIns,
				TTY:         false,
			},
		},
	}
	// 发送请求
	err := streamClient.Send(&gojudge.StreamRequest{Request: req})
	if err != nil {
		return metaerror.Wrap(err, "发送请求失败")
	}

	metaroutine.SafeGo(fmt.Sprintf("runGame-%d-%s", 1, req.RequestID), func() error {
		for {
			// 接收响应
			resp, err := streamClient.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Println("连接已关闭")
					return nil
				}
				return metaerror.Wrap(err, "接收响应失败")
			}
			if resp.Output != nil {
				log.Printf("收到输出: index:%d, fd:%d, %v", resp.Output.Index, resp.Output.Fd, string(resp.Output.Content))
			} else if resp.Response != nil {
				if len(resp.Response.Results) > 0 {
					log.Printf("收到响应: %v", resp.Response.Results[0].String())
					break
				}
			}
		}
		return nil
	})

	// metaroutine.SafeGo(fmt.Sprintf("runGame-%d-%s", 1, req.RequestID), func() error {
	// 	for {
	// 		randValue := metamath.GetRandomInt(1000, 9999)
	// 		slog.Info("发送随机值", "randValue", randValue)
	// 		inputReq := &gojudge.InputRequest{
	// 			Index:   0,
	// 			Fd:      0,
	// 			Content: []byte(fmt.Sprintf("%d\n", randValue)),
	// 		}
	// 		// 发送输入请求
	// 		err = streamClient.Send(&gojudge.StreamRequest{Input: inputReq})
	// 		if err != nil {
	// 			return metaerror.Wrap(err, "发送输入请求失败")
	// 		}
	// 		time.Sleep(3 * time.Second)
	// 	}
	// 	return nil
	// })
	time.Sleep(10 * time.Second)
	return nil
}

// 五子棋评测程序相关函数将在后面定义

// newWebsocket 创建WebSocket连接
func (s *BotService) newWebsocket(args []string, wsURL string) gojudge.Stream {
	header := make(http.Header)
	token := os.Getenv("TOKEN")
	if token != "" {
		header.Add("Authorization", "Bearer "+token)
	}

	// 确保URL是ws://格式
	if len(wsURL) >= 7 && wsURL[:7] == "http://" {
		wsURL = "ws://" + wsURL[7:]
	}

	log.Printf("尝试连接到WebSocket服务器: %s", wsURL)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		log.Printf("WebSocket连接失败: %v", err)
		return nil
	}

	log.Println("WebSocket连接成功")
	return &gojudge.WebsocketStream{Conn: conn}
}
