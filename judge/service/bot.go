package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	foundationbot "foundation/foundation-bot"
	foundationdao "foundation/foundation-dao"
	foundationjudge "foundation/foundation-judge"
	foundationmodel "foundation/foundation-model"
	foundationview "foundation/foundation-view"
	"io"
	botjudge "judge/bot-judge"
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

	maxJob := config.GetConfig().MaxJobBot
	runningCount := int(s.botRunningTasks.Load())
	if runningCount >= maxJob {
		return nil
	}
	ctx := context.Background()
	jobs, err := foundationdao.GetBotReplayDao().RequestBotReplayListPending(
		ctx,
		maxJob-runningCount,
		config.GetConfig().Judger.Key,
	)
	if err != nil {
		return metaerror.Wrap(err, "failed to get run job list")
	}
	jobsCount := len(jobs)
	if jobsCount == 0 {
		return nil
	}

	slog.Info("get bot replay list", "runningCount", runningCount, "maxJob", maxJob, "count", jobsCount)

	s.botRunningTasks.Add(int32(jobsCount))

	for _, job := range jobs {
		metaroutine.SafeGo(
			fmt.Sprintf("RunningBotGame_%d", job.Id), func() error {
				// 执行完本Job后再尝试启动一次任务
				defer s.checkStartBot()

				defer func() {
					slog.Info(fmt.Sprintf("BotGame_%d end", job.Id))
					s.botRunningTasks.Add(-1)
				}()
				val, _ := s.botJobMutexMap.LoadOrStore(job.Id, &botMutexEntry{})
				e := val.(*botMutexEntry)
				atomic.AddInt32(&e.ref, 1)

				defer func() {
					if atomic.AddInt32(&e.ref, -1) == 0 {
						s.botJobMutexMap.Delete(job.Id)
					}
				}()
				e.mu.Lock()
				defer e.mu.Unlock()

				slog.Info(fmt.Sprintf("BotGame_%d start", job.Id))
				err = s.startBotJob(job)
				if err != nil {
					markErr := foundationdao.GetBotReplayDao().MarkBotReplayRunStatus(
						ctx,
						job.Id,
						config.GetConfig().Judger.Key,
						foundationbot.BotGameStatusJudgeFail,
						err.Error(),
					)
					if markErr != nil {
						metapanic.ProcessError(markErr)
					}
					return err
				}

				return nil
			},
		)
	}
	return nil
}

func (s *BotService) startBotJob(job *foundationmodel.BotReplay) error {
	ctx := context.Background()
	codes, err := foundationdao.GetBotCodeDao().GetBotCodes(ctx, job.Bots)
	if err != nil {
		return metaerror.Wrap(err, "failed to get bot replay code map")
	}

	// slog.Info("bot replay code map", "codes", codes)

	codeFiles := make(map[int]map[string]string)
	for _, bc := range codes {
		if !foundationjudge.IsLanguageNeedCompile(bc.Language) {
			continue
		}
		fileIds, compileErr := s.compileBotCode(bc)
		if compileErr != nil {
			return metaerror.Wrap(compileErr, "failed to compile bot code")
		}
		codeFiles[bc.Id] = fileIds
	}
	codeViews := make(map[int]*foundationview.BotCodeView)
	for _, bc := range codes {
		codeViews[bc.Id] = bc
	}

	judgeClient, err := s.runJudgeExec(job)
	if err != nil {
		return metaerror.Wrap(err, "failed to run judge exec")
	}
	defer judgeClient.Close()

	slog.Info("start bot game", "gameId", job.GameId, "bots", job.Bots)

	agents := make(map[int]gojudge.Stream)
	for i, bot := range job.Bots {
		codeView, ok := codeViews[bot]
		if !ok {
			return metaerror.New(fmt.Sprintf("bot %d code not found", bot))
		}
		agent, err := s.runAgent(codeView, codeFiles[bot])
		if err != nil {
			return metaerror.Wrap(err, "failed to run agent")
		}

		// 收到agent的输出后，发送给agent
		metaroutine.SafeGo(fmt.Sprintf("game-%d-agent-%d-receive", job.Id, i), func() error {
			for {
				resp, err := agent.Recv()
				if err != nil {
					if errors.Is(err, io.EOF) {
						log.Printf("agent %d 连接已关闭", i)
						return nil
					}
					return metaerror.Wrap(err, fmt.Sprintf("agent %d 接收响应失败", i))
				}
				if resp.Output != nil {
					paramBytes, err := json.Marshal(botjudge.ChannelContent{
						Index:   i,
						Content: string(resp.Output.Content),
					})
					if err != nil {
						return metaerror.Wrap(err, fmt.Sprintf("agent %d 序列化响应失败", i))
					}
					requestData := botjudge.Request{
						Action: botjudge.ActionTypeOutput,
						Param:  json.RawMessage(paramBytes),
					}
					err = judgeClient.Send(&gojudge.StreamRequest{Input: &gojudge.InputRequest{
						Index:   0,
						Fd:      1,
						Content: []byte(requestData.Json()),
					}})
					if err != nil {
						return metaerror.Wrap(err, fmt.Sprintf("agent %d 发送响应失败", i))
					}
				}
			}
		})

		agents[i] = agent
	}

	metaroutine.SafeGo(fmt.Sprintf("game-%d-judge-receive", job.Id), func() error {
		// JSON缓冲区，用于累积不完整的JSON数据
		jsonBuffer := make([]byte, 0, 4096)

		for {
			// 接收响应
			resp, err := judgeClient.Recv()
			if err != nil {
				if errors.Is(err, io.EOF) {
					log.Println("连接已关闭")
					return nil
				}
				return metaerror.Wrap(err, "接收响应失败")
			}
			if resp.Output != nil {
				// 将新数据添加到缓冲区
				jsonBuffer = append(jsonBuffer, resp.Output.Content...)

				// 处理缓冲区中的JSON数据
				for {
					if len(jsonBuffer) == 0 {
						break
					}

					// 使用botjudge.processJSONBuffer处理缓冲区
					position, err := botjudge.ProcessJSONBuffer(&jsonBuffer, func(req *json.RawMessage) error {
						// 解析请求数据
						var requestData botjudge.Request
						err := json.Unmarshal(*req, &requestData)
						if err != nil {
							return metaerror.Wrap(err, "failed to unmarshal request data")
						}

						// 处理解析出的请求
						if requestData.Action == botjudge.ActionTypeInput {
							var inputReq botjudge.ChannelContent
							err = json.Unmarshal(requestData.Param, &inputReq)
							if err != nil {
								return metaerror.Wrap(err, "failed to unmarshal input request data")
							}
							agent, ok := agents[inputReq.Index]
							if !ok {
								return metaerror.New(fmt.Sprintf("bot %d agent not found", inputReq.Index))
							}
							err = agent.Send(&gojudge.StreamRequest{Input: &gojudge.InputRequest{
								Index:   0,
								Fd:      1,
								Content: []byte(inputReq.Content),
							}})
							if err != nil {
								return metaerror.Wrap(err, "failed to send input request")
							}
						}
						return nil
					})

					if err != nil {
						return metaerror.Wrap(err, "failed to process JSON buffer")
					}

					// 如果没有解析到任何数据，说明缓冲区中的JSON不完整，等待更多数据
					if position == 0 {
						break
					}

					// 移除已处理的部分
					jsonBuffer = jsonBuffer[position:]
				}
			} else if resp.Response != nil {
				if len(resp.Response.Results) > 0 {
					log.Printf("收到响应: %v", resp.Response.Results[0].String())
					break
				}
			}
		}
		return nil
	})

	time.Sleep(10 * time.Second)

	return nil
}

func (s *BotService) runJudgeExec(job *foundationmodel.BotReplay) (gojudge.Stream, error) {

	judgeFileId, err := s.compileJudge(job.GameId)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to compile special judge")
	}
	if judgeFileId == "" {
		return nil, metaerror.New("special judge compile failed")
	}
	wsURL := metahttp.UrlJoin(config.GetConfig().GoJudge.Url, "stream")
	streamClient, err := s.newWebsocket([]string{}, wsURL)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to create websocket")
	}

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
	err = streamClient.Send(&gojudge.StreamRequest{Request: req})
	if err != nil {
		return nil, metaerror.Wrap(err, "发送请求失败")
	}

	return streamClient, nil
}

func (s *BotService) runAgent(codeView *foundationview.BotCodeView, execFileIds map[string]string) (gojudge.Stream, error) {
	wsURL := metahttp.UrlJoin(config.GetConfig().GoJudge.Url, "stream")
	streamClient, err := s.newWebsocket([]string{}, wsURL)
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to create websocket")
	}
	var args []string
	var copyIns map[string]gojudge.CmdFile

	switch codeView.Language {
	case foundationjudge.JudgeLanguageC, foundationjudge.JudgeLanguageCpp,
		foundationjudge.JudgeLanguagePascal, foundationjudge.JudgeLanguageGolang,
		foundationjudge.JudgeLanguageRust:
		args = []string{"a"}
		fileId, ok := execFileIds["a"]
		if !ok {
			return nil, metaerror.New("fileId not found")
		}
		copyIns = map[string]gojudge.CmdFile{
			"a": {
				FileID: fileId,
			},
		}
	case foundationjudge.JudgeLanguageJava:
		className := foundationjudge.GetJavaClass(codeView.Code)
		if className == "" {
			return nil, metaerror.New("class name not found")
		}
		packageName := foundationjudge.GetJavaPackage(codeView.Code)
		qualifiedName := className
		if packageName != "" {
			qualifiedName = packageName + "." + className
		}
		jarFileName := className + ".jar"
		args = []string{
			"java",
			"-Dfile.encoding=UTF-8",
			"-cp",
			jarFileName,
			qualifiedName,
		}
		fileId, ok := execFileIds[jarFileName]
		if !ok {
			return nil, metaerror.New("fileId not found")
		}
		copyIns = map[string]gojudge.CmdFile{
			jarFileName: {
				FileID: fileId,
			},
		}
	case foundationjudge.JudgeLanguagePython:
		args = []string{"python3", "-u", "a.py"}
		copyIns = map[string]gojudge.CmdFile{
			"a.py": {
				Content: codeView.Code,
			},
		}
	case foundationjudge.JudgeLanguageLua:
		args = []string{"luajit", "a.lua"}
		copyIns = map[string]gojudge.CmdFile{
			"a.lua": {
				Content: codeView.Code,
			},
		}
	case foundationjudge.JudgeLanguageTypeScript:
		args = []string{"node", "a.js"}
		fileId, ok := execFileIds["a.js"]
		if !ok {
			return nil, metaerror.New("fileId not found")
		}
		copyIns = map[string]gojudge.CmdFile{
			"a.js": {
				FileID: fileId,
			},
		}
	default:
		return nil, metaerror.New("language not support: %d", codeView.Language)
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
	err = streamClient.Send(&gojudge.StreamRequest{Request: req})
	if err != nil {
		return nil, metaerror.Wrap(err, "发送请求失败")
	}

	return streamClient, nil
}

func (s *BotService) getJudgeFileId(gameId int) string {
	if s.judgeFileIds == nil {
		return ""
	}
	return s.judgeFileIds[gameId]
}

func (s *BotService) compileJudge(gameId int) (string, error) {

	judgeFileId := s.getJudgeFileId(gameId)
	if judgeFileId != "" {
		return judgeFileId, nil
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
	judgeFileId, ok = execFileIds["a"]
	if !ok {
		return "", metaerror.New("special judge compile failed, fileId not found")
	}
	if s.judgeFileIds == nil {
		s.judgeFileIds = make(map[int]string)
	}
	s.judgeFileIds[gameId] = judgeFileId
	return judgeFileId, nil
}

func (s *BotService) compileBotCode(code *foundationview.BotCodeView) (map[string]string, error) {
	runUrl := metahttp.UrlJoin(config.GetConfig().GoJudge.Url, "run")
	execFileIds, extraMessage, compileStatus, err := foundationjudge.CompileCode(
		s.goJudgeClient,
		fmt.Sprintf("bot_%d_code", code.Id),
		runUrl,
		code.Language,
		code.Code,
		GetJudgeService().configFileIds,
		false,
		false,
	)
	if compileStatus != foundationjudge.JudgeStatusAC {
		return nil, metaerror.New("compile bot code failed: %s", extraMessage)
	}
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to compile bot code")
	}
	return execFileIds, nil
}

// newWebsocket 创建WebSocket连接
func (s *BotService) newWebsocket(args []string, wsURL string) (gojudge.Stream, error) {
	header := make(http.Header)
	token := os.Getenv("TOKEN")
	if token != "" {
		header.Add("Authorization", "Bearer "+token)
	}

	// 确保URL是ws://格式
	if len(wsURL) >= 7 && wsURL[:7] == "http://" {
		wsURL = "ws://" + wsURL[7:]
	}
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, header)
	if err != nil {
		return nil, metaerror.Wrap(err, "WebSocket failed to connect")
	}
	return &gojudge.WebsocketStream{Conn: conn}, nil
}
