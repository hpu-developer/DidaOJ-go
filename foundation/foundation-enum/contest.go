package foundationenum

type ContestType int

var (
	ContestTypeAcm ContestType = 0 // ACM模式比赛
	ContestTypeOi  ContestType = 1 // OI模式比赛，最终排名以最后一次提交为准
	ContestTypeIoi ContestType = 2 // IOI模式比赛，以最高分提交为准
)

type ContestScoreType int

var (
	ContestScoreTypeNone     ContestScoreType = 0 // 不启用分数排名，一般用于ACM模式，ACM启用则仅用于展示
	ContestScoreTypeAccepted ContestScoreType = 1 // 题目Accepted才认为得分
	ContestScoreTypePartial  ContestScoreType = 2 // 题目部分得分也按比例得分
)

type ContestDiscussType int

var (
	ContestDiscussTypeNormal  ContestDiscussType = 0 // 正常讨论
	ContestDiscussTypeSelf    ContestDiscussType = 1 // 仅能参与自己发表的讨论（管理员不受限制）
	ContestDiscussTypeDisable ContestDiscussType = 2 // 不接受讨论
)
