package repository

import (
	"time"

	"gin-template/model"

	"gorm.io/gorm"
)

// PackageRepository implements PackageRepository.
type PackageRepository struct {
	db *gorm.DB
}

// NewPackageRepository creates a new instance of PackageRepository.
func NewPackageRepository(db *gorm.DB) *PackageRepository {
	return &PackageRepository{db: db}
}

// GetPackageByID retrieves a package by its ID.
func (r *PackageRepository) GetPackageByID(id uint) (*model.Package, error) {
	var pkg model.Package
	if err := r.db.First(&pkg, id).Error; err != nil {
		return nil, err
	}
	return &pkg, nil
}

// GetAllPackages retrieves all packages.
func (r *PackageRepository) GetAllPackages() ([]model.Package, error) {
	var packages []model.Package
	if err := r.db.Find(&packages).Error; err != nil {
		return nil, err
	}
	return packages, nil
}

// GetUserCurrentSubscription retrieves the current active subscription for a user.
// If no active subscription, it does not return the free package's virtual subscription here.
// The service layer will handle the logic of returning a free package if no active subscription.
func (r *PackageRepository) GetUserCurrentSubscription(userID uint) (*model.Subscription, error) {
	var subscription model.Subscription
	err := r.db.Where("user_id = ? AND status = ? AND expiry_date > ?", userID, "active", time.Now()).
		Order("expiry_date DESC").
		First(&subscription).Error
	if err != nil {
		return nil, err // gorm.ErrRecordNotFound if no active subscription
	}
	return &subscription, nil
}

// GetPackageBySubscription retrieves the package details for a given subscription.
func (r *PackageRepository) GetPackageBySubscription(subscription *model.Subscription) (*model.Package, error) {
	var pkg model.Package
	if err := r.db.First(&pkg, subscription.PackageId).Error; err != nil {
		return nil, err
	}
	return &pkg, nil
}

// CreateSubscription creates a new subscription record in the database.
func (r *PackageRepository) CreateSubscription(userID uint, packageID uint, autoRenew bool, startDate time.Time, expiryDate time.Time, nextRenewalDate time.Time, status string) (*model.Subscription, error) {
	subscription := &model.Subscription{
		UserId:      userID,
		PackageId:   packageID,
		Status:      status, // "active"
		StartDate:   time.Now().Unix(),
		ExpiryDate:  time.Now().Unix(),
		AutoRenew:   autoRenew,
		NextRenewal: time.Now().Unix(),
	}
	if err := r.db.Create(subscription).Error; err != nil {
		return nil, err
	}
	return subscription, nil
}

// UpdateSubscription updates an existing subscription (e.g., to cancel renewal).
func (r *PackageRepository) UpdateSubscription(subscription *model.Subscription) error {
	return r.db.Save(subscription).Error
}

// CreateTokenDistribution records a token distribution event.
func (r *PackageRepository) CreateTokenDistribution(userID uint, subscriptionID uint, packageID uint, amount int, distributedAt time.Time) (*model.TokenDistribution, error) {
	distribution := &model.TokenDistribution{
		UserId:         userID,
		SubscriptionId: subscriptionID,
		PackageId:      packageID,
		Amount:         amount,
		DistributedAt:  time.Now().Unix(),
	}
	if err := r.db.Create(distribution).Error; err != nil {
		return nil, err
	}
	return distribution, nil
}

// InitFreePackage ensures the free package exists in the database.
func (r *PackageRepository) InitFreePackage() error {
	var count int64
	r.db.Model(&model.Package{}).Where("id = ?", model.FreePackage.Id).Count(&count)
	if count == 0 {
		// Use a copy of FreePackage to avoid issues if it's modified elsewhere
		freePkgCopy := model.FreePackage
		return r.db.Create(&freePkgCopy).Error
	}
	return nil
}

// GetFreePackage returns the FreePackage constant.
// This can be used by the service layer.
func (r *PackageRepository) GetFreePackage() *model.Package {
	// Return a copy to prevent modification of the global variable
	freePkgCopy := model.FreePackage
	return &freePkgCopy
}
