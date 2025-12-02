package botjudge

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sync"
)

func Start(startHandle func(*JudgeInfo) error, handle func(*json.RawMessage) error) {
	if handle == nil {
		panic("handle is nil")
	}

	// 重置状态
	stopOnce = sync.Once{}
	stopChan = make(chan struct{})

	runningLock.Lock()
	isRunning = true
	runningLock.Unlock()

	defer func() {
		// 确保在函数结束时更新运行状态
		runningLock.Lock()
		isRunning = false
		runningLock.Unlock()
	}()

	stdinReader := bufio.NewReader(os.Stdin)
	// JSON缓冲区，用于累积不完整的JSON数据
	jsonBuffer := make([]byte, 0, 4096) // 预分配合理的初始容量

	// 创建读取结果通道
	resultChan := make(chan readResult)

	// 启动goroutine进行读取，方便需要的时候可以停止
	go func() {
		for {
			select {
			case <-stopChan:
				return
			default:
			}
			line, err := stdinReader.ReadBytes('\n')
			resultChan <- readResult{line: line, err: err}
		}
	}()

	isFirstHandle := true
	for {
		select {
		case <-stopChan:
			return
		case result := <-resultChan:
			if errors.Is(result.err, io.EOF) {
				return
			}
			if result.err != nil {
				continue
			}
			jsonBuffer = append(jsonBuffer, result.line...)
			position, err := ProcessJSONBuffer(&jsonBuffer, func(data *json.RawMessage) error {
				if isFirstHandle {
					isFirstHandle = false
					var judgeInfo JudgeInfo
					err := json.Unmarshal(*data, &judgeInfo)
					if err != nil {
						return fmt.Errorf("JSON parse: %w", err)
					}
					return startHandle(&judgeInfo)
				}
				return handle(data)
			})
			if err != nil {
				SendError(err)
				return
			}
			if position > 0 {
				jsonBuffer = jsonBuffer[position:]
			}
		}
	}
}
