package controller

import (
	"gin-template/model"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// ProjectRequest 项目请求结构
type ProjectRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Genre       string `json:"genre"`
}

// GetProjects 获取项目列表
func GetProjects(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	
	if page < 1 {
		page = 1
	}
	
	userId := c.GetInt("id")
	
	projects, total, err := model.GetUserProjects(userId, (page-1)*limit, limit)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    projects,
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
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	
	userId := c.GetInt("id")
	username := c.GetString("username")
	
	if projectReq.Title == "" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "项目标题不能为空",
		})
		return
	}
	
	currentTime := time.Now().Format("2006-01-02T15:04:05Z")
	
	project := &model.Project{
		Title:       projectReq.Title,
		Description: projectReq.Description,
		Genre:       projectReq.Genre,
		UserId:      userId,
		Username:    username,
		CreatedAt:   currentTime,
		UpdatedAt:   currentTime,
		LastEditedAt: currentTime,
	}
	
	err := project.Insert()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "项目创建成功",
		"data":    project,
	})
}

// GetProject 获取项目详情
func GetProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	
	userId := c.GetInt("id")
	
	project, err := model.GetProjectById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	// 验证项目所有权
	if project.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权访问该项目",
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    project,
	})
}

// UpdateProject 更新项目信息
func UpdateProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	
	userId := c.GetInt("id")
	
	var projectReq ProjectRequest
	if err := c.ShouldBindJSON(&projectReq); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	
	project, err := model.GetProjectById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	// 验证项目所有权
	if project.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权修改该项目",
		})
		return
	}
	
	// 更新项目信息
	project.Title = projectReq.Title
	project.Description = projectReq.Description
	project.Genre = projectReq.Genre
	project.UpdatedAt = time.Now().Format("2006-01-02T15:04:05Z")
	
	err = project.Update()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "项目更新成功",
		"data":    project,
	})
}

// DeleteProject 删除项目
func DeleteProject(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	
	userId := c.GetInt("id")
	
	project, err := model.GetProjectById(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	// 验证项目所有权
	if project.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权删除该项目",
		})
		return
	}
	
	err = project.Delete()
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "项目删除成功",
	})
} 