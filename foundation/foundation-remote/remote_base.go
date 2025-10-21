package foundationremote

import (
	"context"
	foundationjudge "foundation/foundation-judge"
)

type RemoteAgentBase interface {
	IsSupportJudge(problemId string, language foundationjudge.JudgeLanguage) bool
	PostCrawlProblem(ctx context.Context, id string) (*string, error)
	PostSubmitJudgeJob(
		ctx context.Context,
		problemId string,
		language foundationjudge.JudgeLanguage,
		code string,
	) (string, string, error)
	GetJudgeJobStatus(ctx context.Context, id string) (foundationjudge.JudgeStatus, int, int, int, error)
	GetJudgeJobExtraMessage(ctx context.Context, id string, status foundationjudge.JudgeStatus) (string, error)
}
