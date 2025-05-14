package controller

import (
	"gin-template/define"
	"gin-template/service"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetOutline 获取大纲内容
func GetOutline(c *gin.Context) {
	projectId, _, err := ValidateProjectOwnership(c)
	if err != nil {
		return // 错误已经由ValidateProjectOwnership通过ResponseError返回
	}
	
	// 获取大纲
	outline, err := service.GetOutlineByProjectId(projectId)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOK(c, outline)
}

// SaveOutline 保存/更新大纲内容
func SaveOutline(c *gin.Context) {
	projectId, _, err := ValidateProjectOwnership(c)
	if err != nil {
		return
	}
	
	// 解析请求体
	var outlineReq define.OutlineRequest
	if err := c.ShouldBindJSON(&outlineReq); err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	
	// 保存大纲内容
	outline, err := service.SaveOutlineContent(projectId, outlineReq.Content)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOKWithMessage(c, "大纲保存成功", outline)
}

// GetVersions 获取版本历史
func GetVersions(c *gin.Context) {
	projectId, _, err := ValidateProjectOwnership(c)
	if err != nil {
		return
	}
	
	// 获取版本历史
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	versions, err := service.GetVersionHistory(projectId, limit)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOK(c, versions)
}

// AIGenerate AI续写
func AIGenerate(c *gin.Context) {
	projectId, _, err := ValidateProjectOwnership(c)
	if err != nil {
		return
	}
	
	// 解析请求体
	var aiReq define.AIGenerateRequest
	if err := c.ShouldBindJSON(&aiReq); err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	
	// 调用AI服务生成内容
	result, err := service.GenerateOutlineWithAI(projectId, aiReq.Content, aiReq.Style, aiReq.WordLimit)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOK(c, result)
}

// UploadOutline 上传大纲文件
func UploadOutline(c *gin.Context) {
	// 验证项目所有权（这里projectId未使用，但保留验证是必要的）
	_, _, err := ValidateProjectOwnership(c)
	if err != nil {
		return
	}
	
	// 获取上传的文件
	form, err := c.MultipartForm()
	if err != nil {
		ResponseError(c, "文件上传失败: " + err.Error())
		return
	}
	
	files := form.File["file"]
	if len(files) == 0 {
		ResponseError(c, "未找到上传的文件")
		return
	}
	
	file := files[0] // 只处理第一个文件
	
	// 创建文件头
	fileHeader := &define.FileHeader{
		FileHeader: file,
		SaveFile: func(path string) error {
			return c.SaveUploadedFile(file, path)
		},
	}
	
	// 处理文件上传
	fileInfo, err := service.UploadAndParseOutlineFile(fileHeader)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOK(c, fileInfo)
}

// ExportOutline 导出大纲
func ExportOutline(c *gin.Context) {
	projectId, _, err := ValidateProjectOwnership(c)
	if err != nil {
		return
	}
	
	// 解析请求体
	var exportReq define.ExportRequest
	if err := c.ShouldBindJSON(&exportReq); err != nil {
		ResponseError(c, "无效的参数")
		return
	}
	
	// 导出文件
	result, err := service.ExportOutlineToFile(projectId, exportReq.Format)
	if err != nil {
		ResponseError(c, err.Error())
		return
	}
	
	ResponseOK(c, result)
} 