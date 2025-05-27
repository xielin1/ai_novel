package service

import (
	"encoding/json"
	"errors"
	"fmt"
	"gin-template/define"
	"gin-template/model"
	"gin-template/repository"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PackageService struct {
	packageRepo  *repository.PackageRepository
	tokenService *TokenService
}

func NewPackageService(repo *repository.PackageRepository, tokenService *TokenService) *PackageService {
	return &PackageService{
		packageRepo:  repo,
		tokenService: tokenService,
	}
}

// GetAllPackages 获取所有可用套餐（包括免费套餐）
// 返回：
//
//	define.PackageResponse: 包含套餐信息的响应结构体
//	error: 错误信息（如数据库查询失败、JSON解析失败）
func (s *PackageService) GetAllPackages() (define.PackageResponse, error) {
	// 从仓储层获取所有套餐（不包含免费套餐）
	dbPackages, err := s.packageRepo.GetAllPackages()
	if err != nil {
		return define.PackageResponse{}, fmt.Errorf("failed to get packages from repository: %w", err)
	}

	var packageInfos []define.PackageInfo
	// 遍历数据库套餐，解析JSON格式的功能列表
	for _, pkg := range dbPackages {
		var features []string
		if pkg.Features != "" {
			// 反序列化JSON格式的功能列表，失败时记录日志并忽略
			if err := json.Unmarshal([]byte(pkg.Features), &features); err != nil {
				fmt.Printf("Error unmarshalling features for package %d: %v\n", pkg.Id, err)
				features = []string{}
			}
		}
		packageInfos = append(packageInfos, define.PackageInfo{
			ID:            pkg.Id,
			Name:          pkg.Name,
			Description:   pkg.Description,
			Price:         pkg.Price,
			MonthlyTokens: pkg.MonthlyTokens,
			Duration:      pkg.Duration,
			Features:      features,
		})
	}

	// 检查是否包含免费套餐，若不存在则手动添加
	foundFree := false
	for _, pi := range packageInfos {
		if pi.ID == model.FreePackage.Id {
			foundFree = true
			break
		}
	}
	if !foundFree {
		// 从仓储层获取免费套餐信息
		freePkgModel := s.packageRepo.GetFreePackage()
		var freeFeatures []string
		if freePkgModel.Features != "" {
			// 解析免费套餐的功能列表
			_ = json.Unmarshal([]byte(freePkgModel.Features), &freeFeatures)
		}
		// 将免费套餐插入到结果集头部
		freePackageInfo := define.PackageInfo{
			ID:            freePkgModel.Id,
			Name:          freePkgModel.Name,
			Description:   freePkgModel.Description,
			Price:         freePkgModel.Price,
			MonthlyTokens: freePkgModel.MonthlyTokens,
			Duration:      freePkgModel.Duration,
			Features:      freeFeatures,
		}
		packageInfos = append([]define.PackageInfo{freePackageInfo}, packageInfos...)
	}

	return define.PackageResponse{Packages: packageInfos}, nil
}

// ValidatePackageID 验证套餐ID是否有效
// 参数：
//
//	packageID: 待验证的套餐ID
//
// 返回：
//
//	bool: 是否有效（true有效，false无效）
//	error: 验证过程中发生的错误（如数据库查询失败）
func (s *PackageService) ValidatePackageID(packageID int64) (bool, error) {
	// 免费套餐ID直接视为有效
	if packageID == model.FreePackage.Id {
		return true, nil
	}
	// 通过仓储层查询套餐是否存在
	_, err := s.packageRepo.GetPackageByID(packageID)
	if err != nil {
		// 记录不存在的情况为正常无效，其他错误视为异常
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("error validating package ID %d: %w", packageID, err)
	}
	return true, nil
}

// ValidatePaymentMethod 验证支付方式是否支持
// 参数：
//
//	method: 支付方式字符串（如 alipay/wechat/creditcard）
//
// 返回：是否支持
func ValidatePaymentMethod(method string) bool {
	validPaymentMethods := []string{"alipay", "wechat", "creditcard"}
	for _, m := range validPaymentMethods {
		if method == m {
			return true
		}
	}
	return false
}

// CreateSubscription 创建套餐订阅
// 参数：
//
//	userID: 用户ID
//	req: 创建订阅的请求参数（包含套餐ID、支付方式等）
//
// 返回：
//
//	define.SubscriptionResponse: 订阅结果响应
//	error: 业务逻辑错误（如无效套餐、支付方式不支持、数据库操作失败）
func (s *PackageService) CreateSubscription(userID int64, req define.CreateSubscriptionRequest) (define.SubscriptionResponse, error) {
	// 验证套餐ID有效性
	validPkg, err := s.ValidatePackageID(req.PackageID)
	if err != nil {
		return define.SubscriptionResponse{}, fmt.Errorf("error during package ID validation: %w", err)
	}
	if !validPkg {
		return define.SubscriptionResponse{}, errors.New("invalid package ID")
	}

	// 验证支付方式有效性
	if !ValidatePaymentMethod(req.PaymentMethod) {
		return define.SubscriptionResponse{}, errors.New("unsupported payment method")
	}

	// 获取套餐信息（免费套餐与付费套餐分路径处理）
	var pkgInfo *model.Package
	if req.PackageID == model.FreePackage.Id {
		pkgInfo = s.packageRepo.GetFreePackage()
	} else {
		pkgInfo, err = s.packageRepo.GetPackageByID(req.PackageID)
		if err != nil {
			return define.SubscriptionResponse{}, fmt.Errorf("failed to get package info: %w", err)
		}
	}

	// 计算订阅时间区间
	startDate := time.Now()
	var expiryDate time.Time
	var nextRenewalDate time.Time

	switch pkgInfo.Duration {
	case "monthly":
		expiryDate = startDate.AddDate(0, 1, 0) // 加1个月
	case "yearly":
		expiryDate = startDate.AddDate(1, 0, 0) // 加1年
	case "permanent":
		expiryDate = startDate.AddDate(100, 0, 0) // 永久（设定为100年后过期）
	default:
		return define.SubscriptionResponse{}, fmt.Errorf("unknown package duration: %s", pkgInfo.Duration)
	}

	// 自动续费逻辑：永久套餐不自动续费
	autoRenew := pkgInfo.Duration != "permanent"
	if autoRenew {
		nextRenewalDate = expiryDate // 续费时间为过期日
	}

	// 在仓储层创建订阅记录
	_, err = s.packageRepo.CreateSubscription(userID, req.PackageID, autoRenew, startDate, expiryDate, nextRenewalDate, "active")
	if err != nil {
		return define.SubscriptionResponse{}, fmt.Errorf("failed to create subscription in repository: %w", err)
	}

	// 生成订单ID和交易UUID
	orderID := fmt.Sprintf("ORD%s%04d%04d", time.Now().Format("20060102150405"), userID, req.PackageID)
	transactionUUID := uuid.New().String()

	// 处理代币奖励（仅当存在代币服务且套餐包含代币时）
	var tokenBalance int64
	if s.tokenService != nil && pkgInfo.MonthlyTokens > 0 {
		// 调用代币服务增加用户代币
		userToken, err := s.tokenService.CreditToken(
			userID,
			int64(pkgInfo.MonthlyTokens),
			transactionUUID,
			"package_purchase_credit",
			fmt.Sprintf("购买[%s]套餐奖励", pkgInfo.Name),
			"order",
			orderID,
		)
		if err != nil {
			fmt.Printf("Error crediting tokens for user %d, package %d: %v\n", userID, req.PackageID, err)
		} else if userToken != nil {
			tokenBalance = userToken.Balance // 记录奖励后的余额
		}
	} else if s.tokenService == nil {
		fmt.Printf("TokenService not available, skipping token credit for user %d, package %d\n", userID, req.PackageID)
	}

	// 返回订阅结果
	return define.SubscriptionResponse{
		OrderID:       orderID,
		PackageName:   pkgInfo.Name,
		Amount:        pkgInfo.Price,
		PaymentStatus: "completed",
		ValidUntil:    expiryDate.Format(time.RFC3339), // 过期时间格式化为RFC3339标准
		TokensAwarded: pkgInfo.MonthlyTokens,
		TokenBalance:  tokenBalance,
	}, nil
}

// GetUserCurrentPackageInfo 获取用户当前套餐及订阅信息
// 参数：
//
//	userID: 用户ID
//
// 返回：
//
//	define.CurrentPackageResponse: 当前套餐和订阅详情
//	error: 错误信息（如数据库查询失败）
func (s *PackageService) GetUserCurrentPackageInfo(userID int64) (define.CurrentPackageResponse, error) {
	// 从仓储层获取用户当前订阅
	subscription, err := s.packageRepo.GetUserCurrentSubscription(userID)
	var pkg *model.Package

	if err != nil {
		// 处理订阅不存在的情况，默认返回免费套餐
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg = s.packageRepo.GetFreePackage()
			subscription = &model.Subscription{
				UserId:     userID,
				PackageId:  pkg.Id,
				Status:     "active",
				StartDate:  time.Now().Unix(), // 当前时间作为开始时间
				ExpiryDate: time.Now().Unix(), // 当前时间作为过期时间（免费套餐无实际过期时间）
				AutoRenew:  false,
			}
		} else {
			return define.CurrentPackageResponse{}, fmt.Errorf("error getting user subscription: %w", err)
		}
	} else {
		// 根据订阅获取对应的套餐信息，失败时 fallback 到免费套餐
		pkg, err = s.packageRepo.GetPackageBySubscription(subscription)
		if err != nil {
			fmt.Printf("Error getting package for active subscription %d (user %d): %v. Falling back to free package.\n", subscription.Id, userID, err)
			pkg = s.packageRepo.GetFreePackage()
			subscription.PackageId = pkg.Id // 更新订阅关联的套餐为免费套餐
			subscription.Status = "active"
			subscription.AutoRenew = false
			subscription.NextRenewal = time.Now().Unix() // 下次续费时间设为当前时间
		}
	}

	// 解析套餐功能列表
	var features []string
	if pkg.Features != "" {
		if errJ := json.Unmarshal([]byte(pkg.Features), &features); errJ != nil {
			fmt.Printf("Error unmarshalling features for package %d (current info): %v\n", pkg.Id, errJ)
			features = []string{}
		}
	}

	// 组装响应数据（时间字段使用Unix时间戳）
	respPkg := define.PackageInfo{
		ID:            pkg.Id,
		Name:          pkg.Name,
		Description:   pkg.Description,
		Price:         pkg.Price,
		MonthlyTokens: pkg.MonthlyTokens,
		Duration:      pkg.Duration,
		Features:      features,
	}

	respSubInfo := define.SubscriptionInfo{
		PackageID:   pkg.Id,
		UserID:      userID,
		Status:      subscription.Status,
		StartDate:   subscription.StartDate,  // 订阅开始时间（当前时间）
		ExpiryDate:  subscription.ExpiryDate, // 订阅过期时间（当前时间，免费套餐无实际期限）
		AutoRenew:   subscription.AutoRenew,
		NextRenewal: subscription.NextRenewal, // 下次续费时间
	}

	return define.CurrentPackageResponse{
		Package:            respPkg,
		Subscription:       respSubInfo,
		SubscriptionStatus: subscription.Status,
		StartDate:          subscription.StartDate,  // 订阅开始时间（当前时间）
		ExpiryDate:         subscription.ExpiryDate, // 订阅过期时间（当前时间，免费套餐无实际期限）
		AutoRenew:          subscription.AutoRenew,
		NextRenewalDate:    subscription.NextRenewal,
	}, nil
}

// CancelSubscriptionRenewal 取消订阅自动续费
// 参数：
//
//	userID: 用户ID
//
// 返回：
//
//	define.CancelRenewalResponse: 取消结果响应
//	error: 错误信息（如无有效订阅、数据库更新失败）
func (s *PackageService) CancelSubscriptionRenewal(userID int64) (define.CancelRenewalResponse, error) {
	// 获取用户当前订阅
	subscription, err := s.packageRepo.GetUserCurrentSubscription(userID)
	if err != nil {
		// 处理订阅不存在的情况
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return define.CancelRenewalResponse{}, errors.New("no active subscription found to cancel renewal for")
		}
		return define.CancelRenewalResponse{}, fmt.Errorf("error getting user subscription for cancellation: %w", err)
	}

	// 检查当前是否已开启自动续费
	if !subscription.AutoRenew {
		// 获取套餐名称（用于响应）
		pkgInfo, _ := s.packageRepo.GetPackageByID(subscription.PackageId)
		var pkgName = "N/A"
		if pkgInfo != nil {
			pkgName = pkgInfo.Name
		}
		return define.CancelRenewalResponse{
			PackageName: pkgName,
			ExpiryDate:  time.Now().Unix(), // 过期时间设为当前时间（仅用于提示）
			AutoRenew:   false,
		}, errors.New("auto-renewal is already disabled")
	}

	// 关闭自动续费并更新订阅记录
	subscription.AutoRenew = false
	subscription.NextRenewal = time.Now().Unix() // 清除下次续费时间
	if err := s.packageRepo.UpdateSubscription(subscription); err != nil {
		return define.CancelRenewalResponse{}, fmt.Errorf("failed to update subscription to cancel renewal: %w", err)
	}

	// 获取套餐信息并返回结果
	pkgInfo, err := s.packageRepo.GetPackageBySubscription(subscription)
	if err != nil {
		fmt.Printf("Error getting package info after cancelling renewal for user %d: %v\n", userID, err)
		return define.CancelRenewalResponse{
			PackageName: "N/A",
			ExpiryDate:  time.Now().Unix(),
			AutoRenew:   false,
		}, nil
	}

	return define.CancelRenewalResponse{
		PackageName: pkgInfo.Name,
		ExpiryDate:  time.Now().Unix(),
		AutoRenew:   false,
	}, nil
}

// InitFreePackageInDB 初始化免费套餐到数据库（用于应用启动时确保数据存在）
// 返回：
//
//	error: 初始化过程中的错误（如数据库插入失败）
func (s *PackageService) InitFreePackageInDB() error {
	return s.packageRepo.InitFreePackage()
}
