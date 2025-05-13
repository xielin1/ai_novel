package controller

import (
	"fmt"
	"gin-template/common"
	"gin-template/model"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// OutlineRequest 大纲请求结构
type OutlineRequest struct {
	Content string `json:"content" binding:"required"`
}

// AIGenerateRequest AI续写请求结构
type AIGenerateRequest struct {
	Content   string `json:"content" binding:"required"`
	Style     string `json:"style"`
	WordLimit int    `json:"wordLimit"`
}

// ExportRequest 导出请求结构
type ExportRequest struct {
	Format string `json:"format" binding:"required"`
}

// GetOutline 获取大纲内容
func GetOutline(c *gin.Context) {
	projectId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的项目ID",
		})
		return
	}
	
	// 验证项目所有权
	userId := c.GetInt("id")
	project, err := model.GetProjectById(projectId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "项目不存在",
		})
		return
	}
	
	if project.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权访问该项目",
		})
		return
	}
	
	// 获取大纲
	outline, err := model.GetOutlineByProjectId(projectId)
	if err != nil {
		// 如果是新项目，可能没有大纲
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": gin.H{
				"id": 0,
				"project_id": projectId,
				"content": "",
				"current_version": 0,
				"created_at": "",
				"updated_at": "",
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    outline,
	})
}

// SaveOutline 保存/更新大纲内容
func SaveOutline(c *gin.Context) {
	projectId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的项目ID",
		})
		return
	}
	
	// 验证项目所有权
	userId := c.GetInt("id")
	project, err := model.GetProjectById(projectId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "项目不存在",
		})
		return
	}
	
	if project.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权修改该项目",
		})
		return
	}
	
	// 解析请求体
	var outlineReq OutlineRequest
	if err := c.ShouldBindJSON(&outlineReq); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	
	// 保存大纲内容，创建新版本，非AI生成
	outline, err := model.SaveOutline(projectId, outlineReq.Content, false, "", 0, 0)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	// 更新项目的最后编辑时间
	project.LastEditedAt = time.Now().Format("2006-01-02T15:04:05Z")
	project.Update()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "大纲保存成功",
		"data":    outline,
	})
}

// GetVersions 获取版本历史
func GetVersions(c *gin.Context) {
	projectId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的项目ID",
		})
		return
	}
	
	// 验证项目所有权
	userId := c.GetInt("id")
	project, err := model.GetProjectById(projectId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "项目不存在",
		})
		return
	}
	
	if project.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权访问该项目",
		})
		return
	}
	
	// 获取版本历史
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	if limit <= 0 {
		limit = 10
	}
	
	versions, err := model.GetVersionHistory(projectId, limit)
	if err != nil {
		// 如果是新项目，可能没有版本历史
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []model.Version{},
		})
		return
	}
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    versions,
	})
}

// AIGenerate AI续写
func AIGenerate(c *gin.Context) {
	projectId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的项目ID",
		})
		return
	}
	
	// 验证项目所有权
	userId := c.GetInt("id")
	project, err := model.GetProjectById(projectId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "项目不存在",
		})
		return
	}
	
	if project.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权访问该项目",
		})
		return
	}
	
	// 解析请求体
	var aiReq AIGenerateRequest
	if err := c.ShouldBindJSON(&aiReq); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	
	// 检查Token余额
	requiredTokens := 150 // 假设需要150个token，实际应该根据内容长度计算
	// TODO: 检查用户Token余额
	
	// 调用AI服务进行续写
	// 这里简化处理，实际应调用OpenAI API或其他AI服务
	aiGeneratedContent := "这是AI生成的续写内容示例。实际开发中，这里应该调用OpenAI API或其他AI服务进行续写。\n\n" + 
		"根据用户选择的风格（" + aiReq.Style + "）和字数限制（" + strconv.Itoa(aiReq.WordLimit) + "字），生成相应的续写内容。\n\n" +
		"续写内容将基于原始大纲进行扩展，保持一致性和连贯性，同时根据选定的风格进行适当调整。"
	
	tokensUsed := requiredTokens
	
	// 保存大纲内容，创建新版本，标记为AI生成
	_, err = model.SaveOutline(projectId, aiReq.Content+"\n\n"+aiGeneratedContent, true, aiReq.Style, aiReq.WordLimit, tokensUsed)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	
	// 更新项目的最后编辑时间
	project.LastEditedAt = time.Now().Format("2006-01-02T15:04:05Z")
	project.Update()
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "续写成功",
		"data": gin.H{
			"content":       aiGeneratedContent,
			"tokens_used":   tokensUsed,
			"token_balance": 850, // 假设余额为850
		},
	})
}

// UploadOutline 上传大纲文件
func UploadOutline(c *gin.Context) {
	projectId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的项目ID",
		})
		return
	}
	
	// 验证项目所有权
	userId := c.GetInt("id")
	project, err := model.GetProjectById(projectId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "项目不存在",
		})
		return
	}
	
	if project.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权访问该项目",
		})
		return
	}
	
	// 获取上传的文件
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "文件上传失败: " + err.Error(),
		})
		return
	}
	
	// 检查文件类型
	fileExt := filepath.Ext(file.Filename)
	if fileExt != ".txt" && fileExt != ".docx" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "仅支持.txt和.docx格式的文件",
		})
		return
	}
	
	// 保存文件到临时位置
	tempPath := filepath.Join(common.UploadPath, "temp_"+strconv.Itoa(userId)+"_"+strconv.Itoa(int(time.Now().Unix()))+fileExt)
	if err := c.SaveUploadedFile(file, tempPath); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存文件失败: " + err.Error(),
		})
		return
	}
	
	// 读取文件内容
	var fileContent string
	
	// 简单处理：仅支持txt文件，实际应该支持docx等
	if fileExt == ".txt" {
		data, err := ioutil.ReadFile(tempPath)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "读取文件失败: " + err.Error(),
			})
			return
		}
		fileContent = string(data)
	} else {
		// TODO: 处理docx等其他格式
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "暂不支持docx格式解析",
		})
		return
	}
	
	// 删除临时文件
	os.Remove(tempPath)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "文件上传成功",
		"data": gin.H{
			"content":  fileContent,
			"filename": file.Filename,
			"size":     file.Size,
		},
	})
}

// ExportOutline 导出大纲
func ExportOutline(c *gin.Context) {
	projectId, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的项目ID",
		})
		return
	}
	
	// 验证项目所有权
	userId := c.GetInt("id")
	project, err := model.GetProjectById(projectId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "项目不存在",
		})
		return
	}
	
	if project.UserId != userId {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权访问该项目",
		})
		return
	}
	
	// 解析请求体
	var exportReq ExportRequest
	if err := c.ShouldBindJSON(&exportReq); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无效的参数",
		})
		return
	}
	
	// 检查格式
	if exportReq.Format != "txt" && exportReq.Format != "docx" && exportReq.Format != "pdf" {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "不支持的导出格式，仅支持txt、docx和pdf",
		})
		return
	}
	
	// 获取大纲内容
	outline, err := model.GetOutlineByProjectId(projectId)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "获取大纲内容失败",
		})
		return
	}
	
	// 生成文件名
	fileName := fmt.Sprintf("outline_%d_%s.%s", projectId, time.Now().Format("20060102"), exportReq.Format)
	filePath := filepath.Join(common.UploadPath, fileName)
	
	// 写入文件
	if exportReq.Format == "txt" {
		err = ioutil.WriteFile(filePath, []byte(outline.Content), 0644)
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": "生成文件失败: " + err.Error(),
			})
			return
		}
	} else {
		// TODO: 处理docx和pdf格式导出
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "暂不支持docx和pdf格式导出",
		})
		return
	}
	
	// 返回文件路径
	fileUrl := "/upload/" + fileName
	fileSize := len(outline.Content)
	
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "导出成功",
		"data": gin.H{
			"file_url":  fileUrl,
			"file_size": fileSize,
		},
	})
} 