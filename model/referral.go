package model

import (
	"errors"
	"math/rand"
	"strings"
	"time"
)

// Referral 推荐码模型
type Referral struct {
	Id          uint      `json:"id" gorm:"primaryKey"`
	UserId      uint      `json:"user_id" gorm:"not null;index;unique"`
	Code        string    `json:"code" gorm:"type:varchar(20);not null;unique;index"`
	TotalUsed   int       `json:"total_used" gorm:"default:0"`
	IsActive    bool      `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ReferralUse 推荐码使用记录
type ReferralUse struct {
	Id            uint      `json:"id" gorm:"primaryKey"`
	ReferrerId    uint      `json:"referrer_id" gorm:"not null;index"` // 推荐人ID
	UserId        uint      `json:"user_id" gorm:"not null;index;unique"` // 使用者ID
	ReferralCode  string    `json:"referral_code" gorm:"type:varchar(20);not null"`
	TokensRewarded int      `json:"tokens_rewarded" gorm:"not null"` // 奖励的token数量
	UsedAt        time.Time `json:"used_at" gorm:"not null"`
	CreatedAt     time.Time `json:"created_at"`
}

// GetReferralByUserId 根据用户ID获取推荐码
func GetReferralByUserId(userId uint) (*Referral, error) {
	// 这里实现从数据库查询用户的推荐码
	// 如果不存在，则生成一个新的推荐码
	return nil, nil
}

// GenerateNewReferralCode 为用户生成新的推荐码
func GenerateNewReferralCode(userId uint) (*Referral, error) {
	// 获取用户现有推荐码
	oldReferral, err := GetReferralByUserId(userId)
	if err != nil && err.Error() != "record not found" {
		return nil, err
	}

	// 生成新的随机推荐码
	newCode := generateRandomCode(8)

	// 如果用户之前有推荐码，则更新它
	if oldReferral != nil {
		// oldCode可以用于日志记录或审计
		_ = oldReferral.Code // 标记为已使用避免lint警告
		oldReferral.Code = newCode
		oldReferral.UpdatedAt = time.Now()
		// 更新数据库...
		
		return oldReferral, nil
	}

	// 否则创建一个新的推荐码记录
	newReferral := &Referral{
		UserId:    userId,
		Code:      newCode,
		IsActive:  true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	// 保存到数据库...

	return newReferral, nil
}

// UseReferralCode 使用他人的推荐码
func UseReferralCode(userId uint, referralCode string) (int, error) {
	// 查找推荐码
	var referral Referral
	// 从数据库查询推荐码...

	// 如果找不到推荐码
	if referral.Id == 0 {
		return 0, errors.New("无效的推荐码")
	}

	// 检查推荐码是否为用户自己的
	if referral.UserId == userId {
		return 0, errors.New("不能使用自己的推荐码")
	}

	// 检查用户是否已经使用过推荐码
	var existingUse ReferralUse
	// 从数据库查询是否已经使用过...

	if existingUse.Id > 0 {
		return 0, errors.New("您已经使用过推荐码")
	}

	// 设置奖励的token数量
	tokensRewarded := 200 // 实际应该从系统配置中获取

	// 创建使用记录
	use := ReferralUse{
		ReferrerId:    referral.UserId,
		UserId:        userId,
		ReferralCode:  referralCode,
		TokensRewarded: tokensRewarded,
		UsedAt:        time.Now(),
		CreatedAt:     time.Now(),
	}
	// 保存到数据库...
	_ = use // 标记为已使用避免lint警告

	// 更新推荐人的推荐次数
	referral.TotalUsed++
	// 更新数据库...

	// 给用户增加token
	// 这里应该有更新用户token余额的逻辑...

	return tokensRewarded, nil
}

// GetReferralStat 获取用户的推荐统计信息
func GetReferralStat(userId uint) (int, int, error) {
	// 获取总推荐人数
	// 从数据库统计用户的推荐人数...
	totalReferred := 5 // 示例数据

	// 获取总奖励token数量 
	// 从数据库统计用户通过推荐获得的token总量...
	totalTokensEarned := 1000 // 示例数据

	return totalReferred, totalTokensEarned, nil
}

// GetReferrals 获取用户的推荐记录
func GetReferrals(userId uint, page, limit int) ([]map[string]interface{}, int, error) {
	// 计算偏移量
	offset := (page - 1) * limit
	_ = offset // 标记为已使用避免lint警告

	// 查询推荐记录
	// 从数据库查询用户的推荐记录，并关联用户信息...
	
	// 示例结果
	results := []map[string]interface{}{
		{
			"id":              34,
			"user_id":         156,
			"username":        "user***56",
			"registered_at":   "2023-05-10T09:25:00Z", 
			"tokens_rewarded": 200,
		},
	}

	// 获取总记录数
	totalCount := 5 // 示例数据

	return results, totalCount, nil
}

// 生成随机推荐码
func generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := strings.Builder{}
	result.WriteString("RF")
	
	rand.Seed(time.Now().UnixNano())
	for i := 0; i < length-2; i++ {
		randomIndex := rand.Intn(len(charset))
		result.WriteByte(charset[randomIndex])
	}
	
	return result.String()
} 