package gojudge

type StreamRequest struct {
	Request *RunRequest
	Resize  *ResizeRequest
	Input   *InputRequest
	Cancel  *struct{}
}

type StreamResponse struct {
	Response *RunResponse
	Output   *OutputResponse
}
