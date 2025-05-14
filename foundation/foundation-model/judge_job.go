package foundationmodel

import (
	foundationjudge "foundation/foundation-judge"
	"time"
)

type JudgeJob struct {
	Id int `json:"id" bson:"_id"`

	ProblemId string `json:"problem_id" bson:"problem_id"`                     // 题目ID
	ContestId int    `json:"contest_id,omitempty" bson:"contest_id,omitempty"` // 比赛ID

	Author         int                           `json:"author" bson:"author"`                                       // 提交者UserId
	AuthorUsername *string                       `json:"author_username,omitempty" bson:"author_username,omitempty"` // 申请者用户名
	AuthorNickname *string                       `json:"author_nickname,omitempty" bson:"author_nickname,omitempty"` // 申请者昵称
	ApproveTime    time.Time                     `json:"approve_time" bson:"approve_time"`                           //申请时间
	Language       foundationjudge.JudgeLanguage `json:"language" bson:"language"`                                   // 代码语言
	Code           string                        `json:"code" bson:"code"`                                           // 所评测代码
	CodeLength     int                           `json:"code_length" bson:"code_length"`                             // 代码长度
	Status         foundationjudge.JudgeStatus   `json:"status" bson:"status"`                                       // 评测状态
	JudgeTime      *time.Time                    `json:"judge_time" bson:"judge_time"`                               // 评测的开始时间
	TaskCurrent    int                           `json:"task_current" bson:"task_current"`                           // 评测完成子任务数量
	TaskTotal      int                           `json:"task_total" bson:"task_total"`                               // 评测子任务总数量
	Judger         string                        `json:"judger" bson:"judger"`                                       // 评测机
	Score          int                           `json:"score" bson:"score"`                                         // 所得分数
	Time           int                           `json:"time,omitempty" bson:"time,omitempty"`                       // 所用的时间，纳秒
	Memory         int                           `json:"memory,omitempty" bson:"memory,omitempty"`                   // 所用的内存，byte
	CompileMessage *string                       `json:"compile_message,omitempty" bson:"compile_message,omitempty"` // 编译信息
	Task           []*JudgeTask                  `json:"task,omitempty" bson:"task,omitempty"`                       // 评测子任务
	Private        bool                          `json:"private,omitempty" bson:"private,omitempty"`                 // 是否隐藏源码

	// remote judge 独有信息
	OriginOj        *string `json:"origin_oj,omitempty" bson:"origin_oj,omitempty"`                 // 远程评测OJ
	OriginProblemId *string `json:"origin_problem_id,omitempty" bson:"origin_problem_id,omitempty"` // 远程评测题目ID
	RemoteJudgeId   *string `json:"remote_judge_id,omitempty" bson:"remote_judge_id,omitempty"`     // 远程评测ID
	RemoteAccountId *string `json:"remote_account_id,omitempty" bson:"remote_account_id,omitempty"`
	RemoteLanguage  *string `json:"remote_language,omitempty" bson:"remote_language,omitempty"` // 远程评测语言
}

type JudgeJobBuilder struct {
	item *JudgeJob
}

func NewJudgeJobBuilder() *JudgeJobBuilder {
	return &JudgeJobBuilder{item: &JudgeJob{}}
}

func (b *JudgeJobBuilder) Id(id int) *JudgeJobBuilder {
	b.item.Id = id
	return b
}

func (b *JudgeJobBuilder) ProblemId(problemId string) *JudgeJobBuilder {
	b.item.ProblemId = problemId
	return b
}

func (b *JudgeJobBuilder) Author(author int) *JudgeJobBuilder {
	b.item.Author = author
	return b
}

func (b *JudgeJobBuilder) ApproveTime(approveTime time.Time) *JudgeJobBuilder {
	b.item.ApproveTime = approveTime
	return b
}

func (b *JudgeJobBuilder) Language(language foundationjudge.JudgeLanguage) *JudgeJobBuilder {
	b.item.Language = language
	return b
}

func (b *JudgeJobBuilder) Code(code string) *JudgeJobBuilder {
	b.item.Code = code
	return b
}

func (b *JudgeJobBuilder) CodeLength(codeLength int) *JudgeJobBuilder {
	b.item.CodeLength = codeLength
	return b
}

func (b *JudgeJobBuilder) ContestId(contestId int) *JudgeJobBuilder {
	b.item.ContestId = contestId
	return b
}

func (b *JudgeJobBuilder) Status(status foundationjudge.JudgeStatus) *JudgeJobBuilder {
	b.item.Status = status
	return b
}

func (b *JudgeJobBuilder) JudgeTime(judgeTime *time.Time) *JudgeJobBuilder {
	b.item.JudgeTime = judgeTime
	return b
}

func (b *JudgeJobBuilder) JudgeTaskComplete(judgeTaskComplete int) *JudgeJobBuilder {
	b.item.TaskCurrent = judgeTaskComplete
	return b
}

func (b *JudgeJobBuilder) JudgeTaskTotal(judgeTaskTotal int) *JudgeJobBuilder {
	b.item.TaskTotal = judgeTaskTotal
	return b
}

func (b *JudgeJobBuilder) Score(score int) *JudgeJobBuilder {
	b.item.Score = score
	return b
}

func (b *JudgeJobBuilder) Judger(judger string) *JudgeJobBuilder {
	b.item.Judger = judger
	return b
}

func (b *JudgeJobBuilder) Time(time int) *JudgeJobBuilder {
	b.item.Time = time
	return b
}

func (b *JudgeJobBuilder) Memory(memory int) *JudgeJobBuilder {
	b.item.Memory = memory
	return b
}

func (b *JudgeJobBuilder) CompileMessage(compileMessage *string) *JudgeJobBuilder {
	b.item.CompileMessage = compileMessage
	return b
}

func (b *JudgeJobBuilder) Task(task []*JudgeTask) *JudgeJobBuilder {
	b.item.Task = task
	return b
}

func (b *JudgeJobBuilder) Private(private bool) *JudgeJobBuilder {
	b.item.Private = private
	return b
}

func (b *JudgeJobBuilder) RemoteJudgeId(remoteJudgeId *string) *JudgeJobBuilder {
	b.item.RemoteJudgeId = remoteJudgeId
	return b
}

func (b *JudgeJobBuilder) OriginOj(originOj *string) *JudgeJobBuilder {
	b.item.OriginOj = originOj
	return b
}

func (b *JudgeJobBuilder) OriginProblemId(originProblemId *string) *JudgeJobBuilder {
	b.item.OriginProblemId = originProblemId
	return b
}

func (b *JudgeJobBuilder) RemoteAccountId(remoteAccountId *string) *JudgeJobBuilder {
	b.item.RemoteAccountId = remoteAccountId
	return b
}

func (b *JudgeJobBuilder) RemoteLanguage(remoteLanguage *string) *JudgeJobBuilder {
	b.item.RemoteLanguage = remoteLanguage
	return b
}

func (b *JudgeJobBuilder) Build() *JudgeJob {
	return b.item
}
