package foundationuser

// 金币类型的枚举
type CoinType int

const (
	CoinTypeUnknown  CoinType = 0
	CoinTypeSystem   CoinType = 1
	CoinTypeCheckIn  CoinType = 2
	CoinTypeReward   CoinType = 3
	CoinTypePurchase CoinType = 4
)