package foundationview

import (
	foundationrun "foundation/foundation-run"
)

type RunJob struct {
	Id      int                     `json:"id"`                // 运行任务ID
	Status  foundationrun.RunStatus `json:"status"`            // 运行状态
	Time    int                     `json:"time,omitempty"`    // 运行时间(ms)
	Memory  int                     `json:"memory,omitempty"`  // 运行内存(KB)
	Content string                  `json:"content,omitempty"` // 运行结果输出
}
