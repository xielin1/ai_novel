package repository

import (
	"errors"
	"gin-template/model"
	"time"

	"gorm.io/gorm"
)

type ReferralRepository struct {
	DB *gorm.DB
}

func NewReferralRepository(db *gorm.DB) *ReferralRepository {
	return &ReferralRepository{DB: db}
}

func (r *ReferralRepository) GetReferralCodeByUserId(userId int64) (string, error) {
	var code string
	result := r.DB.Model(&model.Referral{}).
		Where("user_id = ?", userId).
		Pluck("code", &code)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return "", nil // 未找到时返回空字符串
		}
		return "", result.Error
	}
	return code, nil
}

func (r *ReferralRepository) GenerateNewReferralCode(userId int64, newCode string) (*model.Referral, error) {
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

func (r *ReferralRepository) UseReferralCode(userId int64, username string, referralCode string, tokensRewarded int) (int, error) {
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

	// 创建使用记录
	use := model.ReferralUse{
		ReferrerId:     referral.UserId,
		UserId:         userId,
		ReferralCode:   referralCode,
		TokensRewarded: tokensRewarded,
		Username:       username,
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

	return tokensRewarded, nil
}

func (r *ReferralRepository) GetReferralStat(userId int64) (int, int, error) {
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

func (r *ReferralRepository) GetReferrals(userId int64, page, limit int) ([]map[string]interface{}, int, error) {
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
		results = append(results, map[string]interface{}{
			"id":              use.Id,
			"username":        r.maskUsername(use.Username), // 脱敏
			"tokens_rewarded": use.TokensRewarded,
			"used_at":         use.UsedAt,
		})
	}

	return results, int(totalCount), nil
}

func (r *ReferralRepository) maskUsername(username string) string {
	if len(username) <= 3 {
		return username + "***"
	}
	return username[:3] + "***" + username[len(username)-2:]
}
