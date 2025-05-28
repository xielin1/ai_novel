package util

import (
	"fmt"
	"sync"
	"time"

	"github.com/sony/sonyflake"
)

// 业务类型枚举
const (
	BusinessPackage        = "package"
	BusinessReferral       = "referral"
	BusinessReconciliation = "recon"
	BusinessAIWriting      = "ai_writing"
	BusinessInitialBalance = "initial_balance"
)

type UUIDGenerator interface {
	Generate(businessType string) string
}

// 混合生成器（推荐）
type HybridGenerator struct {
	sonyFlake *sonyflake.Sonyflake
	mu        sync.Mutex
}

// 全局 UUID 生成器
var globalUUIDGenerator *HybridGenerator

func GetUUIDGenerator() *HybridGenerator {
	return globalUUIDGenerator
}
func SetUUIDGenerator(uuid *HybridGenerator) {
	globalUUIDGenerator = uuid
}

func NewHybridGenerator(machineID uint16) *HybridGenerator {
	st := sonyflake.Settings{
		MachineID: func() (uint16, error) {
			return machineID, nil
		},
		StartTime: time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
	}
	uuid := &HybridGenerator{
		sonyFlake: sonyflake.NewSonyflake(st),
	}
	SetUUIDGenerator(uuid)
	return uuid
}

func (g *HybridGenerator) Generate(businessType string) string {
	// 组合结构：业务前缀(2位) + 时间戳(32位) + 机器ID(8位) + 序列号(16位)
	uid, _ := g.sonyFlake.NextID()

	return fmt.Sprintf("%s-%d", businessPrefix(businessType), uid)
}

func businessPrefix(businessType string) string {
	switch businessType {
	case BusinessPackage:
		return "PK"
	case BusinessReferral:
		return "RF"
	case BusinessReconciliation:
		return "RC"
	case BusinessAIWriting:
		return "AI"
	default:
		return "DF"
	}
}
