package controller

import (
	"gin-template/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// ProjectRequest 项目请求结构
type ProjectRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Genre       string `json:"genre"`
}

// ProjectController 项目控制器结构体
type ProjectController struct {
	service *service.ProjectService // 注入的服务层实例
}

// NewProjectController 创建项目控制器实例（依赖注入）
func NewProjectController(projectSvc *service.ProjectService) *ProjectController {
	return &ProjectController{
		service: projectSvc,
	}
}

// GetProjects 获取项目列表（结构体方法）
func (c *ProjectController) GetProjects(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	userId := ctx.GetInt("id")

	projects, total, err := c.service.GetUserProjects(userId, (page-1)*limit, limit)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOK(ctx, gin.H{
		"data": projects,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateProject 创建新项目（结构体方法）
func (c *ProjectController) CreateProject(ctx *gin.Context) {
	var projectReq ProjectRequest
	if err := ctx.ShouldBindJSON(&projectReq); err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	userId := ctx.GetInt("id")
	username := ctx.GetString("username")

	if projectReq.Title == "" {
		ResponseError(ctx, "项目标题不能为空")
		return
	}

	project, err := c.service.CreateProject(projectReq.Title, projectReq.Description, projectReq.Genre, userId, username)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOKWithMessage(ctx, "项目创建成功", project)
}

// GetProject 获取项目详情（结构体方法）
func (c *ProjectController) GetProject(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	userId := ctx.GetInt("id")

	project, err := c.service.GetProjectById(id)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	// 验证项目所有权
	if !c.service.CheckProjectOwnership(project, userId) {
		ResponseError(ctx, "无权访问该项目")
		return
	}

	ResponseOK(ctx, project)
}

// UpdateProject 更新项目信息（结构体方法）
func (c *ProjectController) UpdateProject(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	userId := ctx.GetInt("id")

	var projectReq ProjectRequest
	if err := ctx.ShouldBindJSON(&projectReq); err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	project, err := c.service.GetProjectById(id)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	// 验证项目所有权
	if !c.service.CheckProjectOwnership(project, userId) {
		ResponseError(ctx, "无权修改该项目")
		return
	}

	// 更新项目信息
	err = c.service.UpdateProject(project, projectReq.Title, projectReq.Description, projectReq.Genre)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOKWithMessage(ctx, "项目更新成功", project)
}

// DeleteProject 删除项目（结构体方法）
func (c *ProjectController) DeleteProject(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	userId := ctx.GetInt("id")

	project, err := c.service.GetProjectById(id)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	// 验证项目所有权
	if !c.service.CheckProjectOwnership(project, userId) {
		ResponseError(ctx, "无权删除该项目")
		return
	}

	err = c.service.DeleteProject(project)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOKWithMessage(ctx, "项目删除成功", nil)
}
