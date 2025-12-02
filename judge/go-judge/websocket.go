package gojudge

import (
	"encoding/json"
	"fmt"
	"io"
	metaerror "meta/meta-error"

	"github.com/gorilla/websocket"
)

// Stream 定义流接口
type Stream interface {
	Send(req *StreamRequest) error
	Recv() (*StreamResponse, error)
	Close() error
}

// WebsocketStream 结构体定义
type WebsocketStream struct {
	Conn *websocket.Conn
}

// Send 实现Stream接口的Send方法
func (s *WebsocketStream) Send(req *StreamRequest) error {
	w, err := s.Conn.NextWriter(websocket.BinaryMessage)
	if err != nil {
		return metaerror.Wrap(err, "failed to get websocket writer")
	}
	defer w.Close()

	switch {
	case req.Request != nil:
		if _, err := w.Write([]byte{1}); err != nil {
			return err
		}
		if err := json.NewEncoder(w).Encode(req.Request); err != nil {
			return err
		}
	case req.Resize != nil:
		if _, err := w.Write([]byte{2}); err != nil {
			return err
		}
		if err := json.NewEncoder(w).Encode(req.Resize); err != nil {
			return err
		}
	case req.Input != nil:
		index := req.Input.Index & 0xf
		fd := req.Input.Fd & 0xf
		if _, err := w.Write([]byte{3, byte(index<<4 | fd)}); err != nil {
			return err
		}
		if _, err := w.Write(req.Input.Content); err != nil {
			return err
		}
	case req.Cancel != nil:
		if _, err := w.Write([]byte{4}); err != nil {
			return err
		}
	default:
		return fmt.Errorf("invalid request")
	}
	return nil
}

// Recv 实现Stream接口的Recv方法
func (s *WebsocketStream) Recv() (*StreamResponse, error) {
	_, r, err := s.Conn.ReadMessage()
	if err != nil {
		return nil, metaerror.Wrap(err, "failed to read message")
	}
	if len(r) == 0 {
		return nil, io.ErrUnexpectedEOF
	}

	resp := new(StreamResponse)

	// 处理响应消息
	switch r[0] {
	case 1:
		resp.Response = new(RunResponse)
		if err := json.Unmarshal(r[1:], resp.Response); err != nil {
			return nil, metaerror.Wrap(err, "failed to unmarshal run response")
		}
	case 2:
		if len(r) < 2 {
			return nil, io.ErrUnexpectedEOF
		}
		resp.Output = new(OutputResponse)
		resp.Output.Index = int(r[1]>>4) & 0xf
		resp.Output.Fd = int(r[1]) & 0xf
		resp.Output.Content = r[2:]
	default:
		return nil, fmt.Errorf("invalid type code: %d", r[0])
	}
	return resp, nil
}

// Close 实现Stream接口的Close方法
func (s *WebsocketStream) Close() error {
	err := s.Conn.Close()
	if err != nil {
		return metaerror.Wrap(err, "failed to close websocket connection")
	}
	return nil
}
