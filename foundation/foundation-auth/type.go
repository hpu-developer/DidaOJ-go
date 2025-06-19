package foundationauth

type AuthType string

const (
	AuthTypeManageWeb          AuthType = "i-manage-web"
	AuthTypeManageJudge        AuthType = "i-manage-judge"
	AuthTypeManageProblem      AuthType = "i-manage-problem"
	AuthTypeManageProblemDaily AuthType = "i-manage-problem-daily"
	AuthTypeManageContest      AuthType = "i-manage-contest"
	AuthTypeManageCollection   AuthType = "i-manage-collection"
	AuthTypeManageDiscuss      AuthType = "i-manage-discuss"
)
