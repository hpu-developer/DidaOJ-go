package botjudge

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"sync"
)

// 停止控制相关变量
var (
	stopChan    = make(chan struct{})
	stopOnce    sync.Once
	isRunning   bool
	runningLock sync.Mutex
)

func ProcessJSONBuffer(buffer *[]byte, handle func(*json.RawMessage) error) (int, error) {
	if len(*buffer) == 0 {
		return 0, nil
	}

	// 创建一个临时读取器
	reader := bytes.NewReader(*buffer)

	// 创建JSON解码器
	decoder := json.NewDecoder(reader)
	decoder.UseNumber() // 使用Number类型避免精度问题

	// 重置读取器
	reader.Reset(*buffer)

	// 尝试解码JSON对象
	var req json.RawMessage
	err := decoder.Decode(&req)
	if err != nil {
		// 如果是EOF错误，可能是JSON不完整，继续等待更多数据
		if errors.Is(err, io.ErrUnexpectedEOF) {
			return 0, nil
		}

		// 检查是否是语法错误，可能是JSON不完整
		if syntaxErr, ok := err.(*json.SyntaxError); ok {
			// 如果语法错误发生在缓冲区的末尾，可能是JSON不完整
			if int(syntaxErr.Offset) >= len(*buffer)-1 {
				return 0, nil
			}
		}

		return 0, nil
	}

	// 计算已解析的字节数
	position := decoder.InputOffset()
	if position == 0 {
		// 保守处理：如果找不到正确的位置，不跳过任何字符
		// 这样可以确保不会丢失有效数据
		return 0, nil
	}

	// 调用回调函数处理解析后的JSON数据
	if handle != nil {
		err := handle(&req)
		if err != nil {
			return 0, err
		}
	}

	return int(position), nil
}

// Stop 停止JSON处理循环
func Stop() {
	stopOnce.Do(func() {
		runningLock.Lock()
		defer runningLock.Unlock()
		if isRunning {
			close(stopChan)
			isRunning = false
		}
	})
}

// Start 启动JSON处理循环
// 定义读取结果的结构体
type readResult struct {
	line []byte
	err  error
}
