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
	var referral Referral
	result := DB.Where("user_id = ?", userId).First(&referral)
	if result.Error != nil {
		return nil, result.Error
	}
	return &referral, nil
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
		result := DB.Save(oldReferral)
		if result.Error != nil {
			return nil, result.Error
		}
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
	result := DB.Create(newReferral)
	if result.Error != nil {
		return nil, result.Error
	}

	return newReferral, nil
}

// UseReferralCode 使用他人的推荐码
func UseReferralCode(userId uint, referralCode string) (int, error) {
	// 查找推荐码
	var referral Referral
	// 从数据库查询推荐码...
	result := DB.Where("code = ? AND is_active = ?", referralCode, true).First(&referral)
	if result.Error != nil {
		return 0, errors.New("无效的推荐码")
	}

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
	result = DB.Where("user_id = ?", userId).First(&existingUse)
	if result.Error == nil && existingUse.Id > 0 {
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
	result = DB.Create(&use)
	if result.Error != nil {
		return 0, result.Error
	}

	// 更新推荐人的推荐次数
	referral.TotalUsed++
	// 更新数据库...
	result = DB.Save(&referral)
	if result.Error != nil {
		return 0, result.Error
	}

	// 给用户增加token
	// 这里应该有更新用户token余额的逻辑...
	// 这部分将在service层实现

	return tokensRewarded, nil
}

// GetReferralStat 获取用户的推荐统计信息
func GetReferralStat(userId uint) (int, int, error) {
	// 获取总推荐人数
	var totalReferred int64
	result := DB.Model(&ReferralUse{}).Where("referrer_id = ?", userId).Count(&totalReferred)
	if result.Error != nil {
		return 0, 0, result.Error
	}

	// 获取总奖励token数量 
	var totalTokens int
	err := DB.Model(&ReferralUse{}).Where("referrer_id = ?", userId).Select("COALESCE(SUM(tokens_rewarded), 0) as total_tokens").Scan(&totalTokens).Error
	if err != nil {
		return int(totalReferred), 0, err
	}

	return int(totalReferred), totalTokens, nil
}

// GetReferrals 获取用户的推荐记录
func GetReferrals(userId uint, page, limit int) ([]map[string]interface{}, int, error) {
	// 计算偏移量
	offset := (page - 1) * limit

	// 查询推荐记录
	var referrals []ReferralUse
	result := DB.Where("referrer_id = ?", userId).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&referrals)
    
	if result.Error != nil {
		return nil, 0, result.Error
	}
	
	// 获取总记录数
	var totalCount int64
	DB.Model(&ReferralUse{}).Where("referrer_id = ?", userId).Count(&totalCount)
	
	// 转换结果格式，并添加用户信息
	var results []map[string]interface{}
	for _, r := range referrals {
		// 获取被推荐用户信息
		var userObj User
		DB.Select("id, username").Where("id = ?", r.UserId).First(&userObj)
		
		result := map[string]interface{}{
			"id":              r.Id,
			"user_id":         r.UserId,
			"username":        maskUsername(userObj.Username),
			"registered_at":   r.UsedAt.Format(time.RFC3339),
			"tokens_rewarded": r.TokensRewarded,
		}
		results = append(results, result)
	}

	return results, int(totalCount), nil
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

// 对用户名进行脱敏处理
func maskUsername(username string) string {
	if len(username) <= 3 {
		return username + "***"
	}
	
	return username[:3] + "***" + username[len(username)-2:]
} 