package service

import (
	"gin-template/model"
	"time"
)

// ProjectService 处理项目相关的业务逻辑
type ProjectService struct{}

// GetUserProjects 获取用户项目列表
func (s *ProjectService) GetUserProjects(userId, offset, limit int) ([]*model.Project, int64, error) {
	return model.GetUserProjects(userId, offset, limit)
}

// CreateProject 创建新项目
func (s *ProjectService) CreateProject(title, description, genre string, userId int, username string) (*model.Project, error) {
	currentTime := time.Now().Format("2006-01-02T15:04:05Z")
	
	project := &model.Project{
		Title:        title,
		Description:  description,
		Genre:        genre,
		UserId:       userId,
		Username:     username,
		CreatedAt:    currentTime,
		UpdatedAt:    currentTime,
		LastEditedAt: currentTime,
	}
	
	err := project.Insert()
	return project, err
}

// GetProjectById 获取项目详情
func (s *ProjectService) GetProjectById(id int) (*model.Project, error) {
	return model.GetProjectById(id)
}

// UpdateProject 更新项目信息
func (s *ProjectService) UpdateProject(project *model.Project, title, description, genre string) error {
	project.Title = title
	project.Description = description
	project.Genre = genre
	project.UpdatedAt = time.Now().Format("2006-01-02T15:04:05Z")
	
	return project.Update()
}

// DeleteProject 删除项目
func (s *ProjectService) DeleteProject(project *model.Project) error {
	return project.Delete()
}

// CheckProjectOwnership 检查项目所有权
func (s *ProjectService) CheckProjectOwnership(project *model.Project, userId int) bool {
	return project.UserId == userId
} 