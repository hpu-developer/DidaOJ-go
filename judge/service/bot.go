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
	"runtime/debug"
	"strings"
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

// InputRequest defines input operation from the remote
type InputRequest struct {
	Index   int
	Fd      int
	Content []byte
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
	// 五子棋评测程序将在其他地方调用
	// 运行五子棋评测逻辑
	testGomoku()

	return nil
}

// 五子棋评测程序相关函数将在后面定义


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

// GomokuMove 表示五子棋的一步棋
type GomokuMove struct {
	X    int   // X坐标
	Y    int   // Y坐标
	Side int   // 0表示黑棋，1表示白棋
	Err  error // 错误信息
}

// GomokuBoard 表示五子棋棋盘
type GomokuBoard struct {
	Board [][]int // 0表示空，1表示黑棋，2表示白棋
	Size  int     // 棋盘大小
}

// NewGomokuBoard 创建一个新的五子棋棋盘
func NewGomokuBoard(size int) *GomokuBoard {
	board := make([][]int, size)
	for i := range board {
		board[i] = make([]int, size)
	}
	return &GomokuBoard{
		Board: board,
		Size:  size,
	}
}

// MakeMove 在棋盘上落子
func (b *GomokuBoard) MakeMove(x, y, side int) bool {
	// 检查坐标是否有效
	if x < 0 || x >= b.Size || y < 0 || y >= b.Size {
		return false
	}
	// 检查是否已有棋子
	if b.Board[y][x] != 0 {
		return false
	}
	// 落子：黑棋为1，白棋为2
	piece := 1
	if side == 1 {
		piece = 2
	}
	b.Board[y][x] = piece
	return true
}

// CheckWin 检查是否获胜
func (b *GomokuBoard) CheckWin(x, y int) bool {
	piece := b.Board[y][x]
	if piece == 0 {
		return false
	}

	// 检查四个方向：水平、垂直、两个对角线
	directions := [][2]int{
		{1, 0},  // 水平
		{0, 1},  // 垂直
		{1, 1},  // 对角线
		{1, -1}, // 反对角线
	}

	for _, dir := range directions {
		count := 1 // 当前位置已经有一个棋子

		// 正方向
		for i := 1; i < 5; i++ {
			nx, ny := x+i*dir[0], y+i*dir[1]
			if nx >= 0 && nx < b.Size && ny >= 0 && ny < b.Size && b.Board[ny][nx] == piece {
				count++
			} else {
				break
			}
		}

		// 反方向
		for i := 1; i < 5; i++ {
			nx, ny := x-i*dir[0], y-i*dir[1]
			if nx >= 0 && nx < b.Size && ny >= 0 && ny < b.Size && b.Board[ny][nx] == piece {
				count++
			} else {
				break
			}
		}

		// 五子连珠
		if count >= 5 {
			return true
		}
	}

	return false
}

// runGomokuProgram 运行五子棋程序，支持多轮交互
func (s *BotService) runGomokuProgram(side int, myMoveChan chan<- GomokuMove, opponentMoveChan <-chan GomokuMove, gameEndChan chan<- int) {
	wsURL := "http://127.0.0.1:30000/stream"
	sideStr := "白棋"
	if side == 0 {
		sideStr = "黑棋"
	}
	log.Printf("[%s] 开始运行五子棋程序，side=%d", sideStr, side)

	// 初始化棋盘
	board := NewGomokuBoard(15) // 15x15的标准五子棋棋盘

	// 创建持久的WebSocket连接
	streamClient := s.newWebsocket([]string{}, wsURL)
	if streamClient == nil {
		log.Printf("[%s] 创建WebSocket连接失败", sideStr)
		myMoveChan <- GomokuMove{Side: side, Err: errors.New("创建WebSocket连接失败")}
		gameEndChan <- -1
		return
	}
	defer streamClient.Close()

	// 第一次输入：side（0或1）
	firstInput := fmt.Sprintf("%d", side)

	// 创建一个简单的五子棋程序
	// 这里使用echo命令来模拟一个能够处理输入输出的程序
	// 实际使用时应该替换为真实的五子棋AI程序
	// 注意：在真实环境中，应该使用一个长时间运行的程序而不是echo

	// 我们需要为每一步创建一个新的请求，因为echo命令是一次性的
	// 在真实环境中，应该使用一个能够保持运行并处理多轮输入输出的程序

	// 第一步：发送side信息
	sendGomokuRequest(streamClient, side, firstInput)

	// 接收第一步响应
	resp, err := streamClient.Recv()
	if err != nil {
		log.Printf("[%s] 接收side响应失败: %v", sideStr, err)
		myMoveChan <- GomokuMove{Side: side, Err: fmt.Errorf("接收side响应失败: %w", err)}
		gameEndChan <- -1
		return
	}

	// 如果是黑棋，解析并发送第一个坐标
	if side == 0 {
		log.Printf("[%s] 开始落子...", sideStr)
		if len(resp.Results) > 0 && len(resp.Results[0].Files) > 0 {
			if output, ok := resp.Results[0].Files["stdout"]; ok {
				// 按行分割输出
				lines := strings.Split(output, "\n")
				for _, line := range lines {
					line = strings.TrimSpace(line)
					if line == "" {
						continue
					}

					// 尝试解析坐标
					var x, y int
					if _, err := fmt.Sscanf(line, "%d %d", &x, &y); err == nil {
						log.Printf("[%s] 解析到初始坐标: (%d, %d)", sideStr, x, y)

						// 在棋盘上落子
						if board.MakeMove(x, y, side) {
							// 检查是否获胜
							if board.CheckWin(x, y) {
								log.Printf("[%s] 在坐标 (%d, %d) 获胜！", sideStr, x, y)
								gameEndChan <- side
								return
							}
							// 通过通道传递坐标给对手
							myMoveChan <- GomokuMove{X: x, Y: y, Side: side}
							log.Printf("[%s] 已发送初始坐标给对手", sideStr)
							break
						} else {
							log.Printf("[%s] 无效的初始坐标: (%d, %d)", sideStr, x, y)
							myMoveChan <- GomokuMove{X: x, Y: y, Side: side, Err: errors.New("无效的坐标或位置已被占用")}
							gameEndChan <- -1
							return
						}
					}
				}
			}
		}
	}

	// 处理游戏主循环
	done := false
	for !done {
		select {
		case <-time.After(10 * time.Millisecond): // 防止goroutine阻塞
			// 继续循环
			continue
		case opponentMove, ok := <-opponentMoveChan:
			if !ok {
				// 通道已关闭
				done = true
				break
			}

			if opponentMove.Err != nil {
				log.Printf("[%s] 收到对手错误: %v", sideStr, opponentMove.Err)
				gameEndChan <- -1
				done = true
				break
			}

			// 记录对手的落子
			log.Printf("[%s] 收到对手落子: (%d, %d)", sideStr, opponentMove.X, opponentMove.Y)

			// 在自己的棋盘上标记对手的落子
			opponentSide := 1
			if side == 1 {
				opponentSide = 0
			}
			board.MakeMove(opponentMove.X, opponentMove.Y, opponentSide)

			// 发送对手的坐标给当前程序
			moveStr := fmt.Sprintf("%d %d", opponentMove.X, opponentMove.Y)
			sendGomokuRequest(streamClient, side, moveStr)

			// 接收当前程序的响应（自己的落子）
			resp, err := streamClient.Recv()
			if err != nil {
				log.Printf("[%s] 接收落子响应失败: %v", sideStr, err)
				myMoveChan <- GomokuMove{Side: side, Err: fmt.Errorf("接收落子响应失败: %w", err)}
				gameEndChan <- -1
				done = true
				break
			}

			// 解析自己的落子坐标
			if len(resp.Results) > 0 && len(resp.Results[0].Files) > 0 {
				if output, ok := resp.Results[0].Files["stdout"]; ok {
					// 按行分割输出
					lines := strings.Split(output, "\n")
					for _, line := range lines {
						line = strings.TrimSpace(line)
						if line == "" {
							continue
						}

						// 尝试解析坐标
						var x, y int
						if _, err := fmt.Sscanf(line, "%d %d", &x, &y); err == nil {
							log.Printf("[%s] 解析到坐标: (%d, %d)", sideStr, x, y)

							// 在棋盘上落子
							if board.MakeMove(x, y, side) {
								// 检查是否获胜
								if board.CheckWin(x, y) {
									log.Printf("[%s] 在坐标 (%d, %d) 获胜！", sideStr, x, y)
									gameEndChan <- side
									done = true
								}
								// 通过通道传递坐标给对手
								myMoveChan <- GomokuMove{X: x, Y: y, Side: side}
							} else {
								log.Printf("[%s] 无效的坐标: (%d, %d)", sideStr, x, y)
								myMoveChan <- GomokuMove{X: x, Y: y, Side: side, Err: errors.New("无效的坐标或位置已被占用")}
								gameEndChan <- -1
								done = true
								break
							}
						}
					}
				}
			}
		}
	}

	log.Printf("[%s] 程序执行完成", sideStr)
}

// sendGomokuRequest 发送五子棋请求
func sendGomokuRequest(stream Stream, side int, input string) error {
	max10240 := int64(10240)
	stdout := "stdout"
	stderr := "stderr"
	inputFile := "input.txt"

	sideStr := "白棋"
	if side == 0 {
		sideStr = "黑棋"
	}

	// 创建一个简单的程序，能够读取输入并生成响应
	// 这里使用Python来模拟五子棋AI的行为
	pythonCmd := `
import sys

# 读取输入
input_data = sys.stdin.read().strip()

# 如果输入是数字0或1，表示第一次运行，输出身份信息
if input_data == '0' or input_data == '1':
    side = int(input_data)
    print(f"我是{'黑棋' if side == 0 else '白棋'}")
    # 黑棋第一次运行需要立即输出一个坐标
	if side == 0:
		import random
		x = random.randint(0, 14)
		y = random.randint(0, 14)
		print(f"{x} {y}")
	else:
		# 否则认为输入是对手的坐标
		print(f"收到对手坐标: {input_data}")
		# 生成一个随机坐标作为响应
		import random
		x = random.randint(0, 14)
		y = random.randint(0, 14)
		print(f"{x} {y}")
`

	// 创建InputRequest来处理输入操作
	inputReq := &InputRequest{
		Index:   0,  // 第一个输入
		Fd:      0,  // 标准输入文件描述符
		Content: []byte(input),
	}
	
	// 将InputRequest的Content转换为字符串用于CopyIn
	inputContent := string(inputReq.Content)

	req := &Request{
		RequestID: fmt.Sprintf("gomoku-%s-%s-%d", sideStr, time.Now().Format("20060102150405"), time.Now().UnixNano()),
		Cmd: []Cmd{
			{
				Args: []string{"python3", "-c", pythonCmd},
				Env:  []string{"PATH=/usr/bin:/bin"},
				Files: []*CmdFile{
					{Name: &stdout, Max: &max10240},
					{Name: &stderr, Max: &max10240},
				},
				CPULimit:    5000000000, // 5秒
				MemoryLimit: 104857600,  // 100MB
				ProcLimit:   50,
				CopyIn: map[string]CmdFile{
					inputFile: {Content: &inputContent},
				},
				CopyOut: []string{stdout, stderr},
			},
		},
	}

	log.Printf("[%s] 发送请求，输入: %s", sideStr, input)
	return stream.Send(req)
}

// testGomoku 五子棋评测程序的主函数
func testGomoku() error {
	log.Println("===== 开始五子棋评测程序测试 =====")
	defer log.Println("===== 五子棋评测程序测试结束 =====")

	// 创建带缓冲的通道
	blackMoveChan := make(chan GomokuMove, 10) // 带缓冲的通道，避免阻塞
	whiteMoveChan := make(chan GomokuMove, 10)
	gameEndChan := make(chan int, 1) // 用于通知游戏结束：0=黑棋赢，1=白棋赢，-1=平局/错误

	// 使用WaitGroup等待所有goroutine完成
	var wg sync.WaitGroup
	wg.Add(2)

	// 设置超时计时器
	timeout := time.After(5 * time.Minute) // 5分钟超时

	// 启动goroutine运行黑棋程序，添加错误恢复
	go func() {
		defer func() {
			log.Println("黑棋程序goroutine已退出")
			wg.Done()
		}()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("黑棋程序发生panic: %v, 堆栈: %s", r, debug.Stack())
				// 避免向已关闭的通道发送数据
				select {
				case gameEndChan <- -1: // 通知游戏结束
				default:
					// 通道已满或已关闭，忽略
				}
			}
		}()

		log.Println("启动黑棋程序...")
		botService := GetBotService()
		if botService == nil {
			log.Println("获取BotService失败")
			select {
			case gameEndChan <- -1:
			default:
			}
			return
		}
		botService.runGomokuProgram(0, blackMoveChan, whiteMoveChan, gameEndChan)
	}()

	// 启动goroutine运行白棋程序，添加错误恢复
	go func() {
		defer func() {
			log.Println("白棋程序goroutine已退出")
			wg.Done()
		}()
		defer func() {
			if r := recover(); r != nil {
				log.Printf("白棋程序发生panic: %v, 堆栈: %s", r, debug.Stack())
				// 避免向已关闭的通道发送数据
				select {
				case gameEndChan <- -1:
				default:
				}
			}
		}()

		log.Println("启动白棋程序...")
		botService := GetBotService()
		if botService == nil {
			log.Println("获取BotService失败")
			select {
			case gameEndChan <- -1:
			default:
			}
			return
		}
		botService.runGomokuProgram(1, whiteMoveChan, blackMoveChan, gameEndChan)
	}()

	// 主循环，处理游戏逻辑和错误
	winner := -1
	done := false
	moveCount := 0
	maxMoves := 15 * 15 // 棋盘填满即结束

	log.Println("游戏主循环开始")
	for !done {
		select {
		// 处理黑棋的落子
		case move, ok := <-blackMoveChan:
			if !ok {
				log.Println("[主循环] 黑棋通道已关闭")
				done = true
				break
			}
			moveCount++
			log.Printf("[主循环] 收到黑棋落子: (%d, %d), 总步数: %d", move.X, move.Y, moveCount)

			if move.Err != nil {
				log.Printf("[主循环] 黑棋程序错误: %v", move.Err)
				// 通知白棋程序对手出错
				select {
				case whiteMoveChan <- GomokuMove{Side: 0, Err: fmt.Errorf("对手程序错误: %w", move.Err)}:
					log.Println("[主循环] 已通知白棋程序对手错误")
				default:
					log.Println("[主循环] 通知白棋程序失败，通道可能已满")
				}
				winner = -1
				done = true
				break
			}

			// 检查是否达到最大步数
			if moveCount >= maxMoves {
				log.Println("[主循环] 棋盘已满，游戏结束")
				done = true
				break
			}

			// 处理白棋的落子
		case move, ok := <-whiteMoveChan:
			if !ok {
				log.Println("[主循环] 白棋通道已关闭")
				done = true
				break
			}
			moveCount++
			log.Printf("[主循环] 收到白棋落子: (%d, %d), 总步数: %d", move.X, move.Y, moveCount)

			if move.Err != nil {
				log.Printf("[主循环] 白棋程序错误: %v", move.Err)
				// 通知黑棋程序对手出错
				select {
				case blackMoveChan <- GomokuMove{Side: 1, Err: fmt.Errorf("对手程序错误: %w", move.Err)}:
					log.Println("[主循环] 已通知黑棋程序对手错误")
				default:
					log.Println("[主循环] 通知黑棋程序失败，通道可能已满")
				}
				winner = -1
				done = true
				break
			}

			// 检查是否达到最大步数
			if moveCount >= maxMoves {
				log.Println("[主循环] 棋盘已满，游戏结束")
				done = true
				break
			}

		// 处理游戏结束信号
		case result, ok := <-gameEndChan:
			if ok {
				log.Printf("[主循环] 收到游戏结束信号: %d", result)
				winner = result
				done = true
			} else {
				log.Println("[主循环] 游戏结束通道已关闭")
				done = true
			}

		// 处理超时
		case <-timeout:
			log.Println("[主循环] 游戏执行超时")
			winner = -1
			done = true
		}
	}

	log.Println("游戏主循环结束，开始清理资源...")

	// 清理资源，关闭所有通道
	// 注意：需要先关闭输入通道，避免goroutine尝试发送数据到已关闭的通道
	defer func() {
		close(blackMoveChan)
		close(whiteMoveChan)
		// 清空gameEndChan避免阻塞
		select {
		case <-gameEndChan:
		default:
		}
		close(gameEndChan)
		log.Println("所有通道已关闭")
	}()

	// 等待所有goroutine完成
	log.Println("等待goroutine完成...")
	wg.Wait()
	log.Println("所有goroutine已完成")

	// 输出游戏结果
	log.Println("===== 游戏结果 =====")
	if winner == 0 {
		log.Println("黑棋获胜！")
	} else if winner == 1 {
		log.Println("白棋获胜！")
	} else if moveCount >= maxMoves {
		log.Println("棋盘已满，平局！")
	} else {
		log.Println("游戏因错误或超时提前结束")
	}
	log.Printf("总步数: %d", moveCount)
	log.Println("===================")

	return nil
}
