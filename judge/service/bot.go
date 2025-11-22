package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"meta/cron"
	metaerror "meta/meta-error"
	metapanic "meta/meta-panic"
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

	goJudgeClient *http.Client
}

var singletonBotService = singleton.Singleton[BotService]{}

// 我们不再依赖外部的stream和model包，直接在本地定义所需的数据结构

// websocketStream 结构体定义
type websocketStream struct {
	conn *websocket.Conn
}

// CmdFile 表示命令文件
type CmdFile struct {
	Name    *string
	Content *string
	Max     *int64
}

// Cmd 表示执行命令
type Cmd struct {
	Args        []string
	Env         []string
	Files       []*CmdFile
	CPULimit    int64
	MemoryLimit int64
	ProcLimit   int
	CopyIn      map[string]CmdFile
	CopyOut     []string
}

// Request 表示测试请求
type Request struct {
	RequestID string
	Cmd       []Cmd
}

// Result 表示执行结果
type Result struct {
	Status     string
	ExitStatus int
	Files      map[string]string
}

// Response 表示测试响应
type Response struct {
	ErrorMsg string
	Results  []Result
}

// Stream 定义流接口
type Stream interface {
	Send(req *Request) error
	Recv() (*Response, error)
	Close() error
}

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
	s.testBot()

	return nil
}

// 测试机器人连接和请求
func (s *BotService) testBot() {
	// 创建WebSocket连接
	wsURL := "http://127.0.0.1:30000/stream"
	log.Printf("正在连接到 %s", wsURL)
	streamClient := s.newWebsocket([]string{}, wsURL)
	if streamClient == nil {
		log.Println("创建WebSocket连接失败")
		return
	}

	// 发送请求
	resp, err := s.runTest(streamClient)
	if err != nil {
		log.Printf("执行测试失败: %v", err)
		return
	}

	// 打印测试结果
	log.Println("测试执行完成，结果:")
	if resp != nil {
		if resp.ErrorMsg != "" {
			log.Printf("错误信息: %s", resp.ErrorMsg)
		}
		for i, result := range resp.Results {
			log.Printf("命令 %d 状态: %s, 退出码: %d", i, result.Status, result.ExitStatus)
			if len(result.Files) > 0 && result.Files["stdout"] != "" {
				log.Printf("输出内容: %s", result.Files["stdout"])
			}
		}
	} else {
		log.Println("未收到有效响应")
	}
}

// runTest 发送测试请求
func (s *BotService) runTest(sc Stream) (*Response, error) {
	// 构建请求
	content := ""
	max10240 := int64(10240)
	stdout := "stdout"
	stderr := "stderr"

	req := &Request{
		RequestID: "test-request-" + time.Now().Format("20060102150405"),
		Cmd: []Cmd{
			{
				Args: []string{"python3", "-c", "print(123+3)"},
				Env:  []string{"PATH=/usr/bin:/bin"},
				Files: []*CmdFile{
					{Content: &content},
					{Name: &stdout, Max: &max10240},
					{Name: &stderr, Max: &max10240},
				},
				CPULimit:    10000000000, // 10秒
				MemoryLimit: 104857600,   // 100MB
				ProcLimit:   50,
				CopyIn:      make(map[string]CmdFile),
				CopyOut:     []string{"stdout", "stderr"},
			},
		},
	}

	// 发送请求
	log.Printf("正在发送测试请求: %+v", req)
	err := sc.Send(req)
	if err != nil {
		return nil, fmt.Errorf("发送请求失败: %w", err)
	}
	log.Println("测试请求发送成功")

	// 接收响应
	resp, err := sc.Recv()
	if err != nil {
		if errors.Is(err, io.EOF) {
			log.Println("连接已关闭")
			return nil, fmt.Errorf("连接已关闭: %w", err)
		}
		return nil, fmt.Errorf("接收响应失败: %w", err)
	}

	return resp, nil
}

// newWebsocket 创建WebSocket连接
func (s *BotService) newWebsocket(args []string, wsURL string) Stream {
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
	return &websocketStream{conn: conn}
}

// Send 实现Stream接口的Send方法
func (s *websocketStream) Send(req *Request) error {
	log.Println("发送消息到WebSocket服务器")
	w, err := s.conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		log.Printf("获取WebSocket写入器失败: %v", err)
		return err
	}
	defer w.Close()

	// 写入类型标识 1 表示请求
	if _, err := w.Write([]byte{1}); err != nil {
		return err
	}
	if err := json.NewEncoder(w).Encode(req); err != nil {
		log.Printf("编码请求失败: %v", err)
		return err
	}
	return nil
}

// Recv 实现Stream接口的Recv方法
func (s *websocketStream) Recv() (*Response, error) {
	_, r, err := s.conn.ReadMessage()
	if err != nil {
		log.Printf("接收消息失败: %v", err)
		return nil, err
	}
	if len(r) == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	// 处理响应消息
	if r[0] == 1 {
		resp := new(Response)
		if err := json.Unmarshal(r[1:], resp); err != nil {
			log.Printf("解码响应失败: %v", err)
			return nil, err
		}
		return resp, nil
	}
	return nil, fmt.Errorf("无效的类型代码: %d", r[0])
}

// Close 实现Stream接口的Close方法
func (s *websocketStream) Close() error {
	log.Println("关闭WebSocket连接")
	return s.conn.Close()
}
