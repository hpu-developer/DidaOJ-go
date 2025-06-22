package request

import (
	foundationerrorcode "foundation/error-code"
	metaerrorcode "meta/error-code"
)

type ProblemEdit struct {
	Id          string   `json:"id,omitempty" bson:"id,omitempty"`
	Title       string   `json:"title,omitempty" bson:"title,omitempty"`
	Description string   `json:"description,omitempty" bson:"description,omitempty"`
	Source      string   `json:"source,omitempty" bson:"source,omitempty"`
	Private     bool     `json:"private,omitempty" bson:"private,omitempty"`
	TimeLimit   int      `json:"time_limit,omitempty" bson:"time_limit,omitempty"`
	MemoryLimit int      `json:"memory_limit,omitempty" bson:"memory_limit,omitempty"`
	Tags        []string `json:"tags,omitempty" bson:"tags,omitempty"`
}

func (r *ProblemEdit) CheckRequest() (bool, int) {

	if r.Title == "" {
		return false, int(foundationerrorcode.ParamError)
	}
	if r.TimeLimit <= 0 || r.MemoryLimit <= 0 {
		return false, int(foundationerrorcode.ParamError)
	}
	if r.TimeLimit > 30000 {
		return false, int(foundationerrorcode.ParamError)
	}
	if r.MemoryLimit > 1024*1024 {
		return false, int(foundationerrorcode.ParamError)
	}
	return true, int(metaerrorcode.Success)
}
