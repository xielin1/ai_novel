package controller

import (
	"gin-template/define"
	"gin-template/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ProjectController struct {
	service *service.ProjectService
}

func NewProjectController(projectSvc *service.ProjectService) *ProjectController {
	return &ProjectController{
		service: projectSvc,
	}
}

// GetProjects 获取项目列表
func (c *ProjectController) GetProjects(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))

	if page < 1 {
		page = 1
	}

	userId := ctx.GetInt64("id")

	projects, total, err := c.service.GetUserProjects(userId, (page-1)*limit, limit)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOK(ctx, define.BuildPageResponse(projects, total, page, limit))
}

// CreateProject 创建新项目
func (c *ProjectController) CreateProject(ctx *gin.Context) {
	var projectReq define.ProjectRequest
	if err := ctx.ShouldBindJSON(&projectReq); err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	userId := ctx.GetInt64("id")
	username := ctx.GetString("username")

	project, err := c.service.CreateProject(projectReq.Title, projectReq.Description, projectReq.Genre, userId, username)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOKWithMessage(ctx, "项目创建成功", project)
}

// GetProject 获取项目详情
func (c *ProjectController) GetProject(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	userId := ctx.GetInt64("id")

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

// UpdateProject 更新项目信息
func (c *ProjectController) UpdateProject(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	userId := ctx.GetInt64("id")

	var projectReq define.ProjectRequest
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

// DeleteProject 删除项目
func (c *ProjectController) DeleteProject(ctx *gin.Context) {
	id, err := strconv.Atoi(ctx.Param("id"))
	if err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	userId := ctx.GetInt64("id")

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
