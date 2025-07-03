package foundationview

import (
	foundationmodel "foundation/foundation-model"
)

type DiscussCommentList struct {
	foundationmodel.DiscussComment

	InserterUsername string `json:"inserter_username"`
	InserterNickname string `json:"inserter_nickname"`
	ModifierUsername string `json:"modifier_username"`
	ModifierNickname string `json:"modifier_nickname"`
}
