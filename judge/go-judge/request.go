package gojudge

// CmdFile 表示命令文件
type CmdFile struct {
	Src       *string `json:"src"`
	Content   string  `json:"content"`
	FileID    *string `json:"fileId"`
	Name      string  `json:"name"`
	Max       int64   `json:"max"`
	Symlink   *string `json:"symlink"`
	StreamIn  bool    `json:"streamIn"`
	StreamOut bool    `json:"streamOut"`
	Pipe      bool    `json:"pipe"`
}

// Cmd 表示执行命令

type Cmd struct {
	Args  []string   `json:"args"`
	Env   []string   `json:"env,omitempty"`
	Files []*CmdFile `json:"files,omitempty"`

	CPULimit     uint64 `json:"cpuLimit"`
	RealCPULimit uint64 `json:"realCpuLimit"`
	ClockLimit   uint64 `json:"clockLimit"`
	MemoryLimit  uint64 `json:"memoryLimit"`
	StackLimit   uint64 `json:"stackLimit"`
	ProcLimit    uint64 `json:"procLimit"`
	CPURateLimit uint64 `json:"cpuRateLimit"`
	CPUSetLimit  string `json:"cpuSetLimit"`

	CopyIn map[string]CmdFile `json:"copyIn"`

	CopyOut         []string `json:"copyOut"`
	CopyOutCached   []string `json:"copyOutCached"`
	CopyOutMax      uint64   `json:"copyOutMax"`
	CopyOutDir      string   `json:"copyOutDir"`
	CopyOutTruncate bool     `json:"copyOutTruncate"`

	TTY               bool `json:"tty,omitempty"`
	StrictMemoryLimit bool `json:"strictMemoryLimit"`
	DataSegmentLimit  bool `json:"dataSegmentLimit"`
	AddressSpaceLimit bool `json:"addressSpaceLimit"`
}

// PipeIndex defines indexing for a pipe fd
type PipeIndex struct {
	Index int `json:"index"`
	Fd    int `json:"fd"`
}

// PipeMap defines in / out pipe for multiple program
type PipeMap struct {
	In    PipeIndex `json:"in"`
	Out   PipeIndex `json:"out"`
	Name  string    `json:"name"`
	Max   int64     `json:"max"`
	Proxy bool      `json:"proxy"`
}

// RunRequest defines single worker request
type RunRequest struct {
	RequestID   string    `json:"requestId"`
	Cmd         []Cmd     `json:"cmd"`
	PipeMapping []PipeMap `json:"pipeMapping"`
}

// ResizeRequest defines resize operation to the virtual terminal
type ResizeRequest struct {
	Index int `json:"index,omitempty"`
	Fd    int `json:"fd,omitempty"`
	Rows  int `json:"rows,omitempty"`
	Cols  int `json:"cols,omitempty"`
	X     int `json:"x,omitempty"`
	Y     int `json:"y,omitempty"`
}

// InputRequest defines input operation from the remote
type InputRequest struct {
	Index   int
	Fd      int
	Content []byte
}
