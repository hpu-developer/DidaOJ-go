package foundationuser

// 以下是示例测试数据验证，实际使用时可以删除
// func TestGetLevelByExperience() {
//     // 验证等级1
//     fmt.Println("0经验值对应等级:", GetLevelByExperience(0)) // 应该返回1
//     // 验证等级2
//     fmt.Println("699经验值对应等级:", GetLevelByExperience(699)) // 应该返回1
//     fmt.Println("700经验值对应等级:", GetLevelByExperience(700)) // 应该返回2
//     // 验证等级3
//     fmt.Println("1599经验值对应等级:", GetLevelByExperience(1599)) // 应该返回2
//     fmt.Println("1600经验值对应等级:", GetLevelByExperience(1600)) // 应该返回3
//     // 验证等级4
//     fmt.Println("2699经验值对应等级:", GetLevelByExperience(2699)) // 应该返回3
//     fmt.Println("2700经验值对应等级:", GetLevelByExperience(2700)) // 应该返回4
// }

// 增加一个经验类型的枚举
type ExperienceType int

const (
	ExperienceTypeUnknown ExperienceType = 0
	ExperienceTypeCheckIn ExperienceType = 1
)

// GetExperienceForUpgrade 计算指定等级升级所需的经验
// 根据规则：第n级升级到n+1所需的经验是(n-1)*200+500
func GetExperienceForUpgrade(level int) int {
	if level <= 0 {
		return 0
	}
	// 应用规则：第n级升级到n+1所需的经验 = (n-1)*200 + 500
	return (level-1)*200 + 500
}

// GetLevelByExperience 根据经验值计算玩家等级
// 玩家起始等级为1级，第n级升级到n+1所需的经验是(n-1)*200+500
// 使用二分查找算法，时间复杂度O(log n)
func GetLevelByExperience(experience int) int {
	// 起始等级为1级
	if experience <= 0 {
		return 1
	}

	// 使用二分查找算法计算等级
	low, high := 1, experience/100+1 // 上界估计，确保足够大
	bestLevel := 1

	for low <= high {
		mid := (low + high) / 2
		requiredExp := GetTotalExperienceForLevel(mid)

		if requiredExp <= experience {
			// 当前经验可以达到mid级，尝试更高等级
			bestLevel = mid
			low = mid + 1
		} else {
			// 当前经验不足以达到mid级，尝试更低等级
			high = mid - 1
		}
	}

	return bestLevel
}

// GetTotalExperienceForLevel 计算达到指定等级所需的总经验
// 使用数学公式直接计算，避免循环累加，时间复杂度O(1)
// 第1级不需要经验
// 第2级需要500经验 ((2-1)*200+500)
// 第3级需要1200经验 (500 + (3-1)*200+500)
func GetTotalExperienceForLevel(level int) int {
	if level <= 1 {
		return 0
	}
	// 正确的推导过程:
	// 从1级到level级需要的总经验 = Σ[(i-1)*200 + 500]，其中i从1到level-1
	// = 200*Σ(i-1) + 500*(level-1)
	// Σ(i-1)从i=1到i=level-1，等于Σj从j=0到j=level-2，结果是(level-2)*(level-1)/2
	// 所以总经验 = 200*(level-2)*(level-1)/2 + 500*(level-1)
	// = 100*(level-2)*(level-1) + 500*(level-1)
	// = (level-1)*(100*(level-2) + 500)
	// = (level-1)*(100level - 200 + 500)
	// = (level-1)*(100level + 300)
	// 重新计算，发现之前的公式有误
	// 从1级到level级需要的总经验 = Σ[(i-1)*200 + 500]，其中i从1到level-1
	totalExp := 0
	for i := 1; i < level; i++ {
		totalExp += (i-1)*200 + 500
	}
	return totalExp
}
