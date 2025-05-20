package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/model"
	"gin-template/repository"
	"time"
)

// ProjectService 项目服务接口
type ProjectService interface {
	GetUserProjects(userId, offset, limit int) ([]*model.Project, int64, error)
	CreateProject(title, description, genre string, userId int, username string) (*model.Project, error)
	GetProjectById(id int) (*model.Project, error)
	UpdateProject(project *model.Project, title, description, genre string) error
	DeleteProject(project *model.Project) error
	CheckProjectOwnership(project *model.Project, userId int) bool
}

// ProjectServiceImpl 项目服务实现
type ProjectServiceImpl struct {
	repo repository.ProjectRepository // 注入项目仓库
}

// 日志前缀
const projectServiceLogPrefix = "[ProjectService] "

// 全局服务实例
var projectService ProjectService

// SetProjectService 设置全局服务实例
func SetProjectService(service ProjectService) {
	projectService = service
	common.SysLog(projectServiceLogPrefix + "ProjectService已通过依赖注入设置")
}

// GetProjectService 获取全局服务实例
func GetProjectService() ProjectService {
	return projectService
}

// NewProjectService 创建项目服务实例
func NewProjectService(repo repository.ProjectRepository) ProjectService {
	common.SysLog(projectServiceLogPrefix + "初始化ProjectService")
	return &ProjectServiceImpl{repo: repo}
}

// GetUserProjects 获取用户项目列表
func (s *ProjectServiceImpl) GetUserProjects(userId, offset, limit int) ([]*model.Project, int64, error) {
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("获取用户 %d 的项目列表，offset: %d, limit: %d", userId, offset, limit))
	return s.repo.GetUserProjects(userId, offset, limit)
}

// CreateProject 创建新项目
func (s *ProjectServiceImpl) CreateProject(title, description, genre string, userId int, username string) (*model.Project, error) {
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
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("创建新项目: %s，用户Id: %d", title, userId))
	if err := s.repo.CreateProject(project); err != nil {
		common.SysError(projectServiceLogPrefix + fmt.Sprintf("创建项目 %s 失败: %v", title, err))
		return nil, err
	}
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("新项目创建成功，Id: %d", project.Id))
	return project, nil
}

// GetProjectById 获取项目详情
func (s *ProjectServiceImpl) GetProjectById(id int) (*model.Project, error) {
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("获取项目 %d 的详情", id))
	project, err := s.repo.GetProjectById(id)
	if project == nil {
		common.SysLog(projectServiceLogPrefix + fmt.Sprintf("项目 %d 不存在", id))
		return nil, nil
	}
	if err != nil {
		common.SysError(projectServiceLogPrefix + fmt.Sprintf("获取项目 %d 详情失败: %v", id, err))
		return nil, err
	}
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("成功获取项目 %d 的详情", id))
	return project, nil
}

// UpdateProject 更新项目信息
func (s *ProjectServiceImpl) UpdateProject(project *model.Project, title, description, genre string) error {
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("更新项目 %d 的信息", project.Id))
	project.Title = title
	project.Description = description
	project.Genre = genre
	project.UpdatedAt = time.Now().Format("2006-01-02T15:04:05Z")
	if err := s.repo.UpdateProject(project); err != nil {
		common.SysError(projectServiceLogPrefix + fmt.Sprintf("更新项目 %d 失败: %v", project.Id, err))
		return err
	}
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("项目 %d 更新成功", project.Id))
	return nil
}

// DeleteProject 删除项目
func (s *ProjectServiceImpl) DeleteProject(project *model.Project) error {
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("删除项目 %d", project.Id))
	if err := s.repo.DeleteProject(project); err != nil {
		common.SysError(projectServiceLogPrefix + fmt.Sprintf("删除项目 %d 失败: %v", project.Id, err))
		return err
	}
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("项目 %d 删除成功", project.Id))
	return nil
}

// CheckProjectOwnership 检查项目所有权
func (s *ProjectServiceImpl) CheckProjectOwnership(project *model.Project, userId int) bool {
	common.SysLog(projectServiceLogPrefix + fmt.Sprintf("检查用户 %d 是否拥有项目 %d", userId, project.Id))
	result := project.UserId == userId
	if !result {
		common.SysLog(projectServiceLogPrefix + fmt.Sprintf("用户 %d 不拥有项目 %d", userId, project.Id))
	}
	return result
}
