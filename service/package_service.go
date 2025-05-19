package service

import (
	"encoding/json"
	"fmt"
	"gin-template/define"
	"gin-template/model"
	"time"

	"github.com/google/uuid"
)

// 获取所有套餐
func GetAllPackages() (define.PackageResponse, error) {
	// 添加免费版套餐
	freeFeatures := []string{"基础AI续写功能", "每月500个免费Token", "社区支持"}
	freePackage := define.PackageInfo{
		ID:            0,
		Name:          "免费版",
		Description:   "基础功能免费体验",
		Price:         0,
		MonthlyTokens: 500,
		Duration:      "monthly",
		Features:      freeFeatures,
	}
	
	// 这里应从数据库中获取套餐信息
	// 目前使用硬编码的数据，实际应从数据库中查询
	packages := []define.PackageInfo{
		freePackage,
		{
			ID:            1,
			Name:          "基础版",
			Description:   "适合轻度使用的创作者",
			Price:         19.9,
			MonthlyTokens: 5000,
			Duration:      "monthly",
			Features:      []string{"基础AI续写", "历史版本保存"},
		},
		{
			ID:            2,
			Name:          "升级版",
			Description:   "适合中度创作需求",
			Price:         49.9,
			MonthlyTokens: 15000,
			Duration:      "monthly",
			Features:      []string{"高级AI续写", "历史版本保存", "优先客服支持"},
		},
		{
			ID:            3,
			Name:          "永久版会员",
			Description:   "适合专业创作者",
			Price:         199.9,
			MonthlyTokens: 50000,
			Duration:      "permanent",
			Features:      []string{"高级AI续写", "无限历史版本", "专属客服", "高级导出格式"},
		},
	}
	
	return define.PackageResponse{Packages: packages}, nil
}

// 验证套餐ID是否存在
func ValidatePackageID(packageID uint) bool {
	// 免费版特殊处理
	if packageID == 0 {
		return true
	}
	
	// 从数据库查询套餐是否存在
	var count int64
	model.DB.Model(&model.Package{}).Where("id = ?", packageID).Count(&count)
	return count > 0
}

// 验证支付方式
func ValidatePaymentMethod(method string) bool {
	validPaymentMethods := []string{"alipay", "wechat", "creditcard"}
	for _, m := range validPaymentMethods {
		if method == m {
			return true
		}
	}
	return false
}

// 创建订阅
func CreatePackageSubscription(userID uint, request define.CreateSubscriptionRequest) (define.SubscriptionResponse, error) {
	// 验证套餐ID
	if !ValidatePackageID(request.PackageID) {
		return define.SubscriptionResponse{}, fmt.Errorf("无效的套餐ID")
	}
	
	// 验证支付方式
	if !ValidatePaymentMethod(request.PaymentMethod) {
		return define.SubscriptionResponse{}, fmt.Errorf("不支持的支付方式")
	}
	
	// 获取套餐信息
	var packageInfo model.Package
	if request.PackageID != 0 {
		if err := model.DB.First(&packageInfo, request.PackageID).Error; err != nil {
			return define.SubscriptionResponse{}, err
		}
	} else {
		packageInfo = model.FreePackage
	}
	
	// 计算有效期
	validUntil := time.Now().AddDate(0, 1, 0) // 默认一个月
	if packageInfo.Duration == "yearly" {
		validUntil = time.Now().AddDate(1, 0, 0)
	} else if packageInfo.Duration == "permanent" {
		validUntil = time.Now().AddDate(100, 0, 0) // 设置一个很久远的日期
	}
	
	// 创建订阅记录
	subscription := model.Subscription{
		UserId:     userID,
		PackageId:  request.PackageID,
		Status:     "active",
		StartDate:  time.Now(),
		ExpiryDate: validUntil,
		AutoRenew:  packageInfo.Duration != "permanent", // 永久版不自动续费
	}
	
	// 保存订阅记录
	if err := model.DB.Create(&subscription).Error; err != nil {
		return define.SubscriptionResponse{}, err
	}
	
	// 生成订单号
	orderID := fmt.Sprintf("ORD%s%03d", time.Now().Format("20060102"), request.PackageID)
	
	// 生成唯一交易ID，用于幂等性控制
	transactionUUID := uuid.New().String()
	
	// 为用户增加Token余额
	tokensToAward := int64(packageInfo.MonthlyTokens)
	
	// 使用TokenService添加Token
	userToken, err := tokenService.CreditToken(
		userID,
		tokensToAward,
		transactionUUID,
		"package_purchase_credit",
		fmt.Sprintf("购买[%s]套餐奖励", packageInfo.Name),
		"order",
		orderID,
	)
	
	if err != nil {
		// 记录错误，但继续处理，因为用户已经付款
		fmt.Printf("为用户[%d]增加Token失败: %v\n", userID, err)
	}
	
	// 构建响应
	var tokenBalance int64 = 0
	if userToken != nil {
		tokenBalance = userToken.Balance
	}
	
	response := define.SubscriptionResponse{
		OrderID:       orderID,
		PackageName:   packageInfo.Name,
		Amount:        packageInfo.Price,
		PaymentStatus: "completed",
		ValidUntil:    validUntil.Format(time.RFC3339),
		TokensAwarded: packageInfo.MonthlyTokens,
		TokenBalance:  tokenBalance,
	}
	
	return response, nil
}

// 获取用户当前套餐信息
func GetUserCurrentPackageInfo(userID uint) (define.CurrentPackageResponse, error) {
	// 从数据库获取用户的订阅信息
	subscription, packageInfo, err := model.GetUserCurrentPackage(userID)
	if err != nil {
		return define.CurrentPackageResponse{}, err
	}
	
	// 格式化日期为字符串
	startDate := subscription.StartDate.Format(time.RFC3339)
	expiryDate := subscription.ExpiryDate.Format(time.RFC3339)
	
	// 构建响应数据
	var nextRenewalDate string
	if subscription.AutoRenew {
		nextRenewalDate = subscription.ExpiryDate.Format(time.RFC3339)
	}
	
	// 解析features JSON字符串
	var features []string
	if packageInfo.Features != "" {
		json.Unmarshal([]byte(packageInfo.Features), &features)
	}
	
	// 创建PackageInfo
	pkg := define.PackageInfo{
		ID:            packageInfo.Id,
		Name:          packageInfo.Name,
		Description:   packageInfo.Description,
		Price:         packageInfo.Price,
		MonthlyTokens: packageInfo.MonthlyTokens,
		Duration:      packageInfo.Duration,
		Features:      features,
	}
	
	// 创建SubscriptionInfo
	subInfo := define.SubscriptionInfo{
		PackageID:   packageInfo.Id,
		UserID:      userID,
		Status:      subscription.Status,
		StartDate:   startDate,
		ExpiryDate:  expiryDate,
		AutoRenew:   subscription.AutoRenew,
		NextRenewal: nextRenewalDate,
	}
	
	response := define.CurrentPackageResponse{
		Package:            pkg,
		Subscription:       subInfo,
		SubscriptionStatus: subscription.Status,
		StartDate:          startDate,
		ExpiryDate:         expiryDate,
		AutoRenew:          subscription.AutoRenew,
		NextRenewalDate:    nextRenewalDate,
	}
	
	return response, nil
}

// 取消自动续费
func CancelPackageRenewal(userID uint) (define.CancelRenewalResponse, error) {
	// 获取用户当前有效订阅
	var subscription model.Subscription
	result := model.DB.Where("user_id = ? AND status = ? AND expiry_date > ?", 
		userID, "active", time.Now()).
		Order("expiry_date DESC").
		First(&subscription)
	
	if result.Error != nil {
		return define.CancelRenewalResponse{}, result.Error
	}
	
	// 取消自动续费
	subscription.AutoRenew = false
	if err := model.DB.Save(&subscription).Error; err != nil {
		return define.CancelRenewalResponse{}, err
	}
	
	// 获取套餐信息
	var packageInfo model.Package
	if err := model.DB.First(&packageInfo, subscription.PackageId).Error; err != nil {
		return define.CancelRenewalResponse{}, err
	}
	
	response := define.CancelRenewalResponse{
		PackageName: packageInfo.Name,
		ExpiryDate:  subscription.ExpiryDate.Format(time.RFC3339),
		AutoRenew:   false,
	}
	
	return response, nil
} 