package foundationremote

import (
	foundationenum "foundation/foundation-enum"
	"strings"
)

func GetRemoteTypeByString(oj string) foundationenum.RemoteJudgeType {
	oj = strings.ToLower(oj)
	switch oj {
	case "hdu":
		return foundationenum.RemoteJudgeTypeHdu
	case "poj":
		return foundationenum.RemoteJudgeTypePoj
	case "nyoj":
		return foundationenum.RemoteJudgeTypeNyoj
	default:
		return foundationenum.RemoteJudgeTypeLocal
	}
}

func GetRemoteAgent(remoteType foundationenum.RemoteJudgeType) RemoteAgentBase {
	switch remoteType {
	case foundationenum.RemoteJudgeTypeHdu:
		return GetRemoteHduAgent()
	case foundationenum.RemoteJudgeTypePoj:
		return GetRemotePojAgent()
		//case foundationenum.RemoteJudgeTypeNyoj:
		//	return GetRemoteNyojAgent()
	}
	return nil
}
