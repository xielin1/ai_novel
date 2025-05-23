package repository

import (
	"gin-template/model"
	"gorm.io/gorm"
)

// OutlineRepository 提供大纲相关的数据库操作
type OutlineRepository struct {
	DB *gorm.DB
}

// NewOutlineRepository 创建一个新的OutlineRepository实例
func NewOutlineRepository(db *gorm.DB) *OutlineRepository {
	return &OutlineRepository{
		DB: db,
	}
}

// GetOutlineByProjectId 根据项目ID获取大纲
func (r *OutlineRepository) GetOutlineByProjectId(projectId int64) (*model.Outline, error) {
	var outline model.Outline
	err := r.DB.Where("project_id = ?", projectId).First(&outline).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // 返回nil表示未找到记录
		}
		return nil, err
	}
	return &outline, nil
}

// SaveOutline 保存大纲内容并创建新版本
func (r *OutlineRepository) SaveOutline(projectId int64, content string, isAiGenerated bool, aiStyle string, wordLimit int, tokensUsed int) (*model.Outline, error) {

	// 查找是否存在大纲
	var outline model.Outline
	result := r.DB.Where("project_id = ?", projectId).First(&outline)

	if result.Error != nil {
		if result.Error != gorm.ErrRecordNotFound {
			return nil, result.Error
		}

		// 如果不存在，创建新大纲
		outline = model.Outline{
			ProjectId:      projectId,
			Content:        content,
			CurrentVersion: 1,
		}

		if err := r.DB.Create(&outline).Error; err != nil {
			return nil, err
		}
	} else {
		// 如果存在，更新大纲内容和版本
		outline.Content = content
		outline.CurrentVersion += 1

		if err := r.DB.Save(&outline).Error; err != nil {
			return nil, err
		}
	}

	// 创建新版本记录
	version := model.Version{
		OutlineId:     outline.Id,
		VersionNumber: outline.CurrentVersion,
		Content:       content,
		IsAiGenerated: isAiGenerated,
		AiStyle:       aiStyle,
		WordLimit:     wordLimit,
		TokensUsed:    tokensUsed,
	}

	if err := r.DB.Create(&version).Error; err != nil {
		return nil, err
	}

	return &outline, nil
}

// GetVersionHistory 获取版本历史
func (r *OutlineRepository) GetVersionHistory(projectId int64, limit int) ([]*model.Version, error) {
	var outline model.Outline
	var versions []*model.Version

	// 先找到对应的大纲
	err := r.DB.Where("project_id = ?", projectId).First(&outline).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return versions, nil // 返回空列表表示未找到记录
		}
		return nil, err
	}

	// 获取版本历史，按版本号降序排列
	err = r.DB.Where("outline_id = ?", outline.Id).
		Order("version_number desc").
		Limit(limit).
		Find(&versions).Error

	return versions, err
}
