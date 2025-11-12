package foundationuser

import (
	"testing"
)

// TestGetLevelByExperience 测试根据经验值计算玩家等级的函数
func TestGetLevelByExperience(t *testing.T) {
	// 根据规则：第n级升级到n+1所需的经验是(n-1)*200+500
	// 等级1升到2需要500经验
	// 等级2升到3需要700经验
	// 等级3升到4需要900经验
	// 测试用例：经验值 -> 预期等级
	cases := []struct {
		exp      int
		expected int
		name     string
	}{
		{0, 1, "0经验值对应等级1"},
		{499, 1, "499经验值对应等级1（升级前临界值）"},
		{500, 2, "500经验值对应等级2（升级后）"},
		{1199, 2, "1199经验值对应等级2（升级前临界值）"},
		{1200, 3, "1200经验值对应等级3（升级后）"},
		{2099, 3, "2099经验值对应等级3（升级前临界值）"},
		{2100, 4, "2100经验值对应等级4（升级后）"},
		{3199, 4, "3199经验值对应等级4（升级前临界值）"},
		{3200, 5, "3200经验值对应等级5（升级后）"},
		{-10, 1, "负数经验值对应等级1"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetLevelByExperience(tc.exp)
			if result != tc.expected {
				t.Errorf("GetLevelByExperience(%d) = %d; want %d", tc.exp, result, tc.expected)
			}
		})
	}
}

// TestGetExperienceForUpgrade 测试计算指定等级升级所需经验的函数
func TestGetExperienceForUpgrade(t *testing.T) {
	// 测试用例：等级 -> 预期升级所需经验
	cases := []struct {
		level    int
		expected int
		name     string
	}{
		{0, 0, "等级0返回0经验"},
		{1, 500, "等级1升级需要500经验"},
		{2, 700, "等级2升级需要700经验"},
		{3, 900, "等级3升级需要900经验"},
		{4, 1100, "等级4升级需要1100经验"},
		{10, 2300, "等级10升级需要2300经验"},
		{-5, 0, "负数等级返回0经验"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := GetExperienceForUpgrade(tc.level)
			if result != tc.expected {
				t.Errorf("GetExperienceForUpgrade(%d) = %d; want %d", tc.level, result, tc.expected)
			}
		})
	}
}

// TestLevelAndExperienceConsistency 测试等级和经验的一致性
func TestLevelAndExperienceConsistency(t *testing.T) {
	// 测试从1级到10级的一致性
	for level := 1; level <= 10; level++ {
		expRequired := GetExperienceForUpgrade(level)
		currentTotalExp := getTotalExperienceForLevel(level)
		nextLevelTotalExp := getTotalExperienceForLevel(level + 1)

		// 验证升级所需经验等于下一等级总经验减去当前等级总经验
		if nextLevelTotalExp-currentTotalExp != expRequired {
			t.Errorf("等级%d升级逻辑不一致: 计算所需经验=%d, 实际差值=%d",
				level, expRequired, nextLevelTotalExp-currentTotalExp)
		}

		// 验证升级后刚好达到下一级
		nextLevel := GetLevelByExperience(nextLevelTotalExp)
		if nextLevel != level+1 {
			t.Errorf("总经验%d应该对应等级%d，但实际返回等级%d",
				nextLevelTotalExp, level+1, nextLevel)
		}

		// 验证升级前还是当前级
		currentLevel := GetLevelByExperience(nextLevelTotalExp - 1)
		if currentLevel != level {
			t.Errorf("总经验%d应该对应等级%d，但实际返回等级%d",
				nextLevelTotalExp-1, level, currentLevel)
		}
	}
}
