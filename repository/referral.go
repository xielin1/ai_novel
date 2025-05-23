package repository

import (
	"errors"
	"gin-template/model"
	"math/rand"
	"strings"
	"time"

	"gorm.io/gorm"
)

// ReferralRepository 推荐码仓库，处理数据库操作
type ReferralRepository struct {
	DB *gorm.DB
}

// NewReferralRepository 创建推荐码仓库实例
func NewReferralRepository(db *gorm.DB) *ReferralRepository {
	return &ReferralRepository{DB: db}
}

// GetReferralByUserId 根据用户ID获取推荐码
func (r *ReferralRepository) GetReferralByUserId(userId uint) (*model.Referral, error) {
	var referral model.Referral
	result := r.DB.Where("user_id = ?", userId).First(&referral)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, nil // 未找到时返回 nil
		}
		return nil, result.Error
	}
	return &referral, nil
}

// GenerateNewReferralCode 生成新推荐码
func (r *ReferralRepository) GenerateNewReferralCode(userId uint) (*model.Referral, error) {
	newCode := r.generateRandomCode(8) // 私有方法生成随机码

	newReferral := &model.Referral{
		UserId:   userId,
		Code:     newCode,
		IsActive: true,
	}
	if err := r.DB.Create(newReferral).Error; err != nil {
		return nil, err
	}
	return newReferral, nil
}

// UseReferralCode 使用推荐码
func (r *ReferralRepository) UseReferralCode(userId uint, referralCode string) (int, error) {
	// 查找有效推荐码
	var referral model.Referral
	result := r.DB.Where("code = ? AND is_active = ?", referralCode, true).First(&referral)
	if result.Error != nil || referral.Id == 0 {
		return 0, errors.New("无效的推荐码")
	}

	// 检查是否为自己的推荐码
	if referral.UserId == userId {
		return 0, errors.New("不能使用自己的推荐码")
	}

	// 检查用户是否已使用过推荐码
	var existingUse model.ReferralUse
	if err := r.DB.Where("user_id = ?", userId).First(&existingUse).Error; err == nil {
		return 0, errors.New("您已经使用过推荐码")
	}

	// 设置奖励（实际应从配置获取，此处硬编码示例）
	tokensRewarded := 200

	// 创建使用记录
	use := model.ReferralUse{
		ReferrerId:     referral.UserId,
		UserId:         userId,
		ReferralCode:   referralCode,
		TokensRewarded: tokensRewarded,
		UsedAt:         time.Now().Unix(),
	}
	if err := r.DB.Create(&use).Error; err != nil {
		return 0, err
	}

	// 更新推荐人使用次数
	referral.TotalUsed++
	if err := r.DB.Save(&referral).Error; err != nil {
		return 0, err
	}

	return tokensRewarded, nil // 返回奖励的Token数量（如需返回用户余额，需调用TokenService）
}

// GetReferralStat 获取推荐统计信息
func (r *ReferralRepository) GetReferralStat(userId uint) (int, int, error) {
	var totalReferred int64
	if err := r.DB.Model(&model.ReferralUse{}).Where("referrer_id = ?", userId).Count(&totalReferred).Error; err != nil {
		return 0, 0, err
	}

	var totalTokens int
	if err := r.DB.Model(&model.ReferralUse{}).
		Where("referrer_id = ?", userId).
		Select("COALESCE(SUM(tokens_rewarded), 0)").
		Scan(&totalTokens).Error; err != nil {
		return int(totalReferred), 0, err
	}

	return int(totalReferred), totalTokens, nil
}

// GetReferrals 获取推荐记录列表
func (r *ReferralRepository) GetReferrals(userId uint, page, limit int) ([]map[string]interface{}, int, error) {
	offset := (page - 1) * limit
	var referrals []model.ReferralUse
	result := r.DB.Where("referrer_id = ?", userId).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&referrals)
	if result.Error != nil {
		return nil, 0, result.Error
	}

	// 统计总数
	var totalCount int64
	if err := r.DB.Model(&model.ReferralUse{}).Where("referrer_id = ?", userId).Count(&totalCount).Error; err != nil {
		return nil, 0, err
	}

	// 转换结果并脱敏用户名
	var results []map[string]interface{}
	for _, use := range referrals {
		var user struct {
			ID       uint
			Username string
		}
		if err := r.DB.Select("id, username").Where("id = ?", use.UserId).First(&user).Error; err != nil {
			return nil, 0, err
		}

		results = append(results, map[string]interface{}{
			"id":              use.Id,
			"user_id":         use.UserId,
			"username":        r.maskUsername(user.Username), // 私有方法脱敏
			"tokens_rewarded": use.TokensRewarded,
		})
	}

	return results, int(totalCount), nil
}

// generateRandomCode 生成随机推荐码（私有方法）
func (r *ReferralRepository) generateRandomCode(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	rand.Seed(time.Now().UnixNano())
	var b strings.Builder
	b.WriteString("RF") // 前缀
	for i := 0; i < length-2; i++ {
		b.WriteByte(charset[rand.Intn(len(charset))])
	}
	return b.String()
}

// maskUsername 用户名脱敏（私有方法）
func (r *ReferralRepository) maskUsername(username string) string {
	if len(username) <= 3 {
		return username + "***"
	}
	return username[:3] + "***" + username[len(username)-2:]
}
