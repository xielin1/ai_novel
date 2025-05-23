package repository

import (
	"gin-template/model"
	"gorm.io/gorm"
)

// projectRepository 项目仓库实现
type ProjectRepository struct {
	db *gorm.DB
}

// NewProjectRepository 创建项目仓库实例
func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

// GetUserProjects 获取用户项目列表
func (r *ProjectRepository) GetUserProjects(userId int64, offset, limit int) ([]*model.Project, int64, error) {
	var projects []*model.Project
	var total int64
	if err := r.db.Model(&model.Project{}).Where("user_id = ?", userId).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := r.db.Where("user_id = ?", userId).Order("id desc").Limit(limit).Offset(offset).Find(&projects).Error; err != nil {
		return nil, 0, err
	}
	return projects, total, nil
}

// CreateProject 创建新项目
func (r *ProjectRepository) CreateProject(project *model.Project) error {
	return r.db.Create(project).Error
}

// GetProjectById 获取项目详情
func (r *ProjectRepository) GetProjectById(id int) (*model.Project, error) {
	var project model.Project
	err := r.db.Where("id = ?", id).First(&project).Error
	if err == gorm.ErrRecordNotFound {
		return nil, nil // 未找到时返回 nil
	}
	return &project, err
}

// UpdateProject 更新项目信息
func (r *ProjectRepository) UpdateProject(project *model.Project) error {
	return r.db.Save(project).Error
}

// DeleteProject 删除项目
func (r *ProjectRepository) DeleteProject(project *model.Project) error {
	return r.db.Delete(project).Error
}
