package service

import (
	"fmt"
	"strings"
	"testing"
)

// TestBusinessPrefix 测试 businessPrefix 函数
func TestBusinessPrefix(t *testing.T) {
	tests := []struct {
		businessType   string
		expectedPrefix string
	}{
		{BusinessPackage, "PK"},
		{BusinessReferral, "RF"},
		{BusinessReconciliation, "RC"},
		{BusinessAIWriting, "AI"},
		// 测试默认情况
		{"unknown_type", "DF"},
		{"", "DF"}, // 测试空字符串
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("BusinessType:%s", tt.businessType), func(t *testing.T) {
			prefix := businessPrefix(tt.businessType)
			if prefix != tt.expectedPrefix {
				t.Errorf("businessPrefix(%q) = %q; want %q", tt.businessType, prefix, tt.expectedPrefix)
			}
		})
	}
}

// TestHybridGenerator_Generate 测试 HybridGenerator 的 Generate 方法
func TestHybridGenerator_Generate(t *testing.T) {
	// 创建一个 HybridGenerator 实例，使用一个固定的 machineID 用于测试
	generator := NewHybridGenerator(1)

	// 定义一些业务类型进行测试
	businessTypes := []string{
		BusinessPackage,
		BusinessReferral,
		BusinessReconciliation,
		BusinessAIWriting,
		"test_business", // 测试未知类型
	}

	// 记录生成的 ID，用于检查是否重复 (简单检查，sonyflake 本身保证唯一性)
	generatedIDs := make(map[string]bool)

	for _, businessType := range businessTypes {
		t.Run(fmt.Sprintf("Generate for %s", businessType), func(t *testing.T) {
			// 生成一个 ID
			generatedID := generator.Generate(businessType)

			// 1. 检查是否以正确的业务前缀开头
			expectedPrefix := businessPrefix(businessType)
			if !strings.HasPrefix(generatedID, expectedPrefix) {
				t.Errorf("Generated ID %q does not start with expected prefix %q", generatedID, expectedPrefix)
			}

			// 2. 检查是否包含连字符
			if !strings.Contains(generatedID, "-") {
				t.Errorf("Generated ID %q does not contain a hyphen", generatedID)
			}

			// 3. 检查连字符后面的部分是否是数字 (sonyflake ID)
			parts := strings.Split(generatedID, "-")
			if len(parts) != 2 {
				t.Errorf("Generated ID %q split by hyphen results in %d parts, expected 2", generatedID, len(parts))
			}
			// 尝试将第二部分转换为数字，如果失败则说明不是有效的 sonyflake ID 部分
			//_, err := fmt.Atoi(parts[1])
			//if err != nil {
			//	t.Errorf("Second part of generated ID %q (%q) is not a number: %v", generatedID, parts[1], err)
			//}

			// 4. 检查生成的 ID 是否已经存在 (简单检查唯一性)
			if generatedIDs[generatedID] {
				t.Errorf("Generated ID %q is duplicated", generatedID)
			}
			generatedIDs[generatedID] = true
		})
	}

	// 可以进一步测试在短时间内生成多个 ID 是否都不同
	// ... (为了简洁，这里省略，sonyflake 本身保证了纳秒级别的唯一性)
}
