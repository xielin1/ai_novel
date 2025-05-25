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

// NewPackageService creates a new instance of PackageService.
// Note: TokenService needs to be passed as an argument once its interface is defined/available.
func NewPackageService(repo *repository.PackageRepository, tokenService *TokenService) *PackageService {
	return &PackageService{
		packageRepo:  repo,
		tokenService: tokenService,
	}
}

// GetAllPackages retrieves all available packages, including the free tier.
func (s *PackageService) GetAllPackages() (define.PackageResponse, error) {
	dbPackages, err := s.packageRepo.GetAllPackages()
	if err != nil {
		return define.PackageResponse{}, fmt.Errorf("failed to get packages from repository: %w", err)
	}

	var packageInfos []define.PackageInfo
	for _, pkg := range dbPackages {
		var features []string
		if pkg.Features != "" {
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

	foundFree := false
	for _, pi := range packageInfos {
		if pi.ID == model.FreePackage.Id {
			foundFree = true
			break
		}
	}
	if !foundFree {
		freePkgModel := s.packageRepo.GetFreePackage()
		var freeFeatures []string
		if freePkgModel.Features != "" {
			_ = json.Unmarshal([]byte(freePkgModel.Features), &freeFeatures)
		}
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

// ValidatePackageID checks if a package ID is valid.
func (s *PackageService) ValidatePackageID(packageID int64) (bool, error) {
	if packageID == model.FreePackage.Id {
		return true, nil
	}
	_, err := s.packageRepo.GetPackageByID(packageID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("error validating package ID %d: %w", packageID, err)
	}
	return true, nil
}

// ValidatePaymentMethod (can remain a helper or be part of a payment service if more complex)
func ValidatePaymentMethod(method string) bool {
	validPaymentMethods := []string{"alipay", "wechat", "creditcard"}
	for _, m := range validPaymentMethods {
		if method == m {
			return true
		}
	}
	return false
}

// CreateSubscription handles the business logic for creating a new package subscription.
func (s *PackageService) CreateSubscription(userID int64, req define.CreateSubscriptionRequest) (define.SubscriptionResponse, error) {
	validPkg, err := s.ValidatePackageID(req.PackageID)
	if err != nil {
		return define.SubscriptionResponse{}, fmt.Errorf("error during package ID validation: %w", err)
	}
	if !validPkg {
		return define.SubscriptionResponse{}, errors.New("invalid package ID")
	}

	if !ValidatePaymentMethod(req.PaymentMethod) {
		return define.SubscriptionResponse{}, errors.New("unsupported payment method")
	}

	var pkgInfo *model.Package
	if req.PackageID == model.FreePackage.Id {
		pkgInfo = s.packageRepo.GetFreePackage()
	} else {
		pkgInfo, err = s.packageRepo.GetPackageByID(req.PackageID)
		if err != nil {
			return define.SubscriptionResponse{}, fmt.Errorf("failed to get package info: %w", err)
		}
	}

	startDate := time.Now()
	var expiryDate time.Time
	var nextRenewalDate time.Time

	switch pkgInfo.Duration {
	case "monthly":
		expiryDate = startDate.AddDate(0, 1, 0)
	case "yearly":
		expiryDate = startDate.AddDate(1, 0, 0)
	case "permanent":
		expiryDate = startDate.AddDate(100, 0, 0)
	default:
		return define.SubscriptionResponse{}, fmt.Errorf("unknown package duration: %s", pkgInfo.Duration)
	}

	autoRenew := pkgInfo.Duration != "permanent"
	if autoRenew {
		nextRenewalDate = expiryDate
	}

	_, err = s.packageRepo.CreateSubscription(userID, req.PackageID, autoRenew, startDate, expiryDate, nextRenewalDate, "active")
	if err != nil {
		return define.SubscriptionResponse{}, fmt.Errorf("failed to create subscription in repository: %w", err)
	}

	orderID := fmt.Sprintf("ORD%s%04d%04d", time.Now().Format("20060102150405"), userID, req.PackageID)
	transactionUUID := uuid.New().String()

	var tokenBalance int64
	if s.tokenService != nil && pkgInfo.MonthlyTokens > 0 {
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
			tokenBalance = userToken.Balance
		}
	} else if s.tokenService == nil {
		fmt.Printf("TokenService not available, skipping token credit for user %d, package %d\n", userID, req.PackageID)
	}

	return define.SubscriptionResponse{
		OrderID:       orderID,
		PackageName:   pkgInfo.Name,
		Amount:        pkgInfo.Price,
		PaymentStatus: "completed",
		ValidUntil:    expiryDate.Format(time.RFC3339),
		TokensAwarded: pkgInfo.MonthlyTokens,
		TokenBalance:  tokenBalance,
	}, nil
}

// GetUserCurrentPackageInfo retrieves the user's current package and subscription details.
func (s *PackageService) GetUserCurrentPackageInfo(userID int64) (define.CurrentPackageResponse, error) {
	subscription, err := s.packageRepo.GetUserCurrentSubscription(userID)
	var pkg *model.Package

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			pkg = s.packageRepo.GetFreePackage()
			subscription = &model.Subscription{
				UserId:     userID,
				PackageId:  pkg.Id,
				Status:     "active",
				StartDate:  time.Now().Unix(),
				ExpiryDate: time.Now().Unix(),
				AutoRenew:  false,
			}
		} else {
			return define.CurrentPackageResponse{}, fmt.Errorf("error getting user subscription: %w", err)
		}
	} else {
		pkg, err = s.packageRepo.GetPackageBySubscription(subscription)
		if err != nil {
			fmt.Printf("Error getting package for active subscription %d (user %d): %v. Falling back to free package.\n", subscription.Id, userID, err)
			pkg = s.packageRepo.GetFreePackage()
			subscription.PackageId = pkg.Id
			subscription.Status = "active"
			subscription.AutoRenew = false
			subscription.NextRenewal = time.Now().Unix()
		}
	}

	var features []string
	if pkg.Features != "" {
		if errJ := json.Unmarshal([]byte(pkg.Features), &features); errJ != nil {
			fmt.Printf("Error unmarshalling features for package %d (current info): %v\n", pkg.Id, errJ)
			features = []string{}
		}
	}

	startDateStr := time.Now().Unix()
	expiryDateStr := time.Now().Unix()
	var nextRenewalDateStr int64
	if subscription.AutoRenew {
		nextRenewalDateStr = time.Now().Unix()
	}

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
		StartDate:   startDateStr,
		ExpiryDate:  expiryDateStr,
		AutoRenew:   subscription.AutoRenew,
		NextRenewal: nextRenewalDateStr,
	}

	return define.CurrentPackageResponse{
		Package:            respPkg,
		Subscription:       respSubInfo,
		SubscriptionStatus: subscription.Status,
		StartDate:          startDateStr,
		ExpiryDate:         expiryDateStr,
		AutoRenew:          subscription.AutoRenew,
		NextRenewalDate:    nextRenewalDateStr,
	}, nil
}

// CancelSubscriptionRenewal cancels the auto-renewal for a user's active subscription.
func (s *PackageService) CancelSubscriptionRenewal(userID int64) (define.CancelRenewalResponse, error) {
	subscription, err := s.packageRepo.GetUserCurrentSubscription(userID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return define.CancelRenewalResponse{}, errors.New("no active subscription found to cancel renewal for")
		}
		return define.CancelRenewalResponse{}, fmt.Errorf("error getting user subscription for cancellation: %w", err)
	}

	if !subscription.AutoRenew {
		pkgInfo, _ := s.packageRepo.GetPackageByID(subscription.PackageId)
		var pkgName = "N/A"
		if pkgInfo != nil {
			pkgName = pkgInfo.Name
		}
		return define.CancelRenewalResponse{
			PackageName: pkgName,
			ExpiryDate:  time.Now().Unix(),
			AutoRenew:   false,
		}, errors.New("auto-renewal is already disabled")
	}

	subscription.AutoRenew = false
	subscription.NextRenewal = time.Now().Unix()
	if err := s.packageRepo.UpdateSubscription(subscription); err != nil {
		return define.CancelRenewalResponse{}, fmt.Errorf("failed to update subscription to cancel renewal: %w", err)
	}

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

// InitFreePackageInDB is a utility function that could be called during application startup
// to ensure the free package is in the database.
func (s *PackageService) InitFreePackageInDB() error {
	return s.packageRepo.InitFreePackage()
}
