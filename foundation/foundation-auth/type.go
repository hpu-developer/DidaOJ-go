package foundationauth

type AuthType string

const (
	AuthTypeManageJudge      AuthType = "i-manage-judge"
	AuthTypeManageProblem    AuthType = "i-manage-problem"
	AuthTypeManageContest    AuthType = "i-manage-contest"
	AuthTypeManageCollection AuthType = "i-manage-collection"
)
