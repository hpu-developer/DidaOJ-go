package request

type ProblemEdit struct {
	Id          string   `json:"id,omitempty" bson:"id,omitempty"`
	Title       string   `json:"title,omitempty" bson:"title,omitempty"`
	Description string   `json:"description,omitempty" bson:"description,omitempty"`
	Source      string   `json:"source,omitempty" bson:"source,omitempty"`
	TimeLimit   int      `json:"time_limit,omitempty" bson:"time_limit,omitempty"`
	MemoryLimit int      `json:"memory_limit,omitempty" bson:"memory_limit,omitempty"`
	Tags        []string `json:"tags,omitempty" bson:"tags,omitempty"`
}
