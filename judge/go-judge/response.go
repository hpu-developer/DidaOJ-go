package gojudge

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Result defines single command result
type Result struct {
	Status     Status            `json:"status"`
	ExitStatus int               `json:"exitStatus"`
	Error      string            `json:"error,omitempty"`
	Time       uint64            `json:"time"`
	Memory     uint64            `json:"memory"`
	RunTime    uint64            `json:"runTime"`
	ProcPeak   uint64            `json:"procPeak,omitempty"`
	Files      map[string]string `json:"files,omitempty"`
	FileIDs    map[string]string `json:"fileIds,omitempty"`

	files []string
	Buffs map[string][]byte `json:"-"`
}

func (r Result) String() string {
	type Result struct {
		Status     Status
		ExitStatus int
		Error      string
		Time       time.Duration
		RunTime    time.Duration
		ProcPeak   uint64
		Files      map[string]string
		FileIDs    map[string]string
	}
	d := Result{
		Status:     r.Status,
		ExitStatus: r.ExitStatus,
		Error:      r.Error,
		Time:       time.Duration(r.Time),
		RunTime:    time.Duration(r.RunTime),
		ProcPeak:   r.ProcPeak,
		Files:      make(map[string]string),
		FileIDs:    r.FileIDs,
	}
	for k, v := range r.Files {
		d.Files[k] = "len:" + strconv.Itoa(len(v))
	}
	return fmt.Sprintf("%+v", d)
}

// Close need to be called when mmap specified to be true
func (r *Result) Close() {
	// remove temporary files
	for _, f := range r.files {
		os.Remove(f)
	}
	// remove potential mmap
	for _, b := range r.Buffs {
		releaseByte(b)
	}
}

func releaseByte(b []byte) {
}

// RunResponse defines worker response for single request
type RunResponse struct {
	RequestID string   `json:"requestId"`
	Results   []Result `json:"results"`
	ErrorMsg  string   `json:"error,omitempty"`

	mmap bool
}

// Close need to be called when mmap specified to be true
func (r *RunResponse) Close() {
	if !r.mmap {
		return
	}
	for _, res := range r.Results {
		res.Close()
	}
}

type OutputResponse struct {
	Index   int
	Fd      int
	Content []byte
}
