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

var projectService = &service.ProjectService{}

// GetProjects 获取项目列表
func GetProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	if page < 1 {
		page = 1
	}
	
	userId := c.GetInt("id")
	
	projects, total, err := projectService.GetUserProjects(userId, (page-1)*limit, limit)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOK(c, gin.H{
		"data": projects,
		"pagination": gin.H{
			"total": total,
			"page":  page,
			"limit": limit,
			"pages": (total + int64(limit) - 1) / int64(limit),
		},
	})
}

// CreateProject 创建新项目
func CreateProject(c *gin.Context) {
	var projectReq ProjectRequest
	if err := c.ShouldBindJSON(&projectReq); err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	
	userId := c.GetInt("id")
	username := c.GetString("username")
	
	if projectReq.Title == "" {
		ResponseError(c, "项目标题不能为空")
		return
	}
	
	project, err := projectService.CreateProject(projectReq.Title, projectReq.Description, projectReq.Genre, userId, username)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOKWithMessage(c, "项目创建成功", project)
}

// GetProject 获取项目详情
func GetProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	
	userId := c.GetInt("id")
	
	project, err := projectService.GetProjectById(id)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	// 验证项目所有权
	if !projectService.CheckProjectOwnership(project, userId) {
		ResponseError(c, "无权访问该项目")
		return
	}
	
	ResponseOK(c, project)
}

// UpdateProject 更新项目信息
func UpdateProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	
	userId := c.GetInt("id")
	
	var projectReq ProjectRequest
	if err := c.ShouldBindJSON(&projectReq); err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	
	project, err := projectService.GetProjectById(id)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	// 验证项目所有权
	if !projectService.CheckProjectOwnership(project, userId) {
		ResponseError(c, "无权修改该项目")
		return
	}
	
	// 更新项目信息
	err = projectService.UpdateProject(project, projectReq.Title, projectReq.Description, projectReq.Genre)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOKWithMessage(c, "项目更新成功", project)
}

// DeleteProject 删除项目
func DeleteProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	
	userId := c.GetInt("id")
	
	project, err := projectService.GetProjectById(id)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	// 验证项目所有权
	if !projectService.CheckProjectOwnership(project, userId) {
		ResponseError(c, "无权删除该项目")
		return
	}
	
	err = projectService.DeleteProject(project)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOKWithMessage(c, "项目删除成功", nil)
} 