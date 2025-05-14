package controller

import (
	"gin-template/model"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Response 统一响应结构
type Response struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

// ResponseOK 返回成功响应
func ResponseOK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: "",
		Data:    data,
	})
}

// ResponseOKWithMessage 返回带消息的成功响应
func ResponseOKWithMessage(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// ResponseError 返回错误响应
func ResponseError(c *gin.Context, message string) {
	c.JSON(http.StatusOK, Response{
		Success: false,
		Message: message,
		Data:    nil,
	})
}

// ResponseErrorWithStatus 返回带状态码的错误响应
func ResponseErrorWithStatus(c *gin.Context, status int, message string) {
	c.JSON(status, Response{
		Success: false,
		Message: message,
		Data:    nil,
	})
}

// ResponseErrorWithData 返回带数据的错误响应
func ResponseErrorWithData(c *gin.Context, message string, data interface{}) {
	c.JSON(http.StatusOK, Response{
		Success: false,
		Message: message,
		Data:    data,
	})
}

// ValidateProjectOwnership 验证项目所有权
// 返回项目ID、项目信息和错误（如有）
func ValidateProjectOwnership(c *gin.Context) (int, *model.Project, error) {
	// 解析项目ID
	projectId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		ResponseError(c, "无效的项目ID")
		return 0, nil, err
	}
	
	// 获取当前用户ID
	userId := c.GetInt("id")
	
	// 获取项目信息
	project, err := model.GetProjectById(projectId)
	if err != nil {
		ResponseError(c, "项目不存在")
		return projectId, nil, err
	}
	
	// 验证所有权
	if project.UserId != userId {
		ResponseError(c, "无权访问该项目")
		return projectId, project, err
	}
	
	return projectId, project, nil
} 