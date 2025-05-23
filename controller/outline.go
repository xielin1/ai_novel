package controller

import (
	"gin-template/define"
	"gin-template/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

// OutlineController 大纲控制器结构体
type OutlineController struct {
	service *service.OutlineService // 注入的服务层实例
}

// NewOutlineController 创建大纲控制器实例（依赖注入）
func NewOutlineController(outlineSvc *service.OutlineService) *OutlineController {
	return &OutlineController{
		service: outlineSvc,
	}
}

// GetOutline 获取大纲内容（结构体方法）
func (c *OutlineController) GetOutline(ctx *gin.Context) {
	projectId, _, err := ValidateProjectOwnership(ctx)
	if err != nil {
		return // 错误处理由中间件或验证函数完成
	}

	outline, err := c.service.GetOutlineByProjectId(projectId) // 使用注入的服务实例
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOK(ctx, outline)
}

// SaveOutline 保存/更新大纲内容（结构体方法）
func (c *OutlineController) SaveOutline(ctx *gin.Context) {
	projectId, _, err := ValidateProjectOwnership(ctx)
	if err != nil {
		return
	}

	var outlineReq define.OutlineRequest
	if err := ctx.ShouldBindJSON(&outlineReq); err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	outline, err := c.service.SaveOutlineContent(projectId, outlineReq.Content) // 使用注入的服务实例
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOKWithMessage(ctx, "大纲保存成功", outline)
}

// GetVersions 获取版本历史（结构体方法）
func (c *OutlineController) GetVersions(ctx *gin.Context) {
	projectId, _, err := ValidateProjectOwnership(ctx)
	if err != nil {
		return
	}

	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "10"))
	versions, err := c.service.GetVersionHistory(projectId, limit) // 使用注入的服务实例
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOK(ctx, versions)
}

// AIGenerate AI续写（结构体方法）
func (c *OutlineController) AIGenerate(ctx *gin.Context) {
	projectId, project, err := ValidateProjectOwnership(ctx) // 假设验证函数返回用户ID
	if err != nil {
		return
	}

	var aiReq define.AIGenerateRequest
	if err := ctx.ShouldBindJSON(&aiReq); err != nil {
		ResponseError(ctx, "无效的参数")
		return
	}

	result, err := c.service.GenerateOutlineWithAI( // 使用注入的服务实例
		project.UserId,
		projectId,
		aiReq.Content,
		aiReq.Style,
		aiReq.WordLimit,
	)
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOK(ctx, result)
}

// ParseOutline 解析大纲文件（结构体方法）
func (c *OutlineController) ParseOutline(ctx *gin.Context) {
	// 验证项目所有权（保留必要验证）
	_, _, err := ValidateProjectOwnership(ctx)
	if err != nil {
		return
	}

	form, err := ctx.MultipartForm()
	if err != nil {
		ResponseError(ctx, "文件上传失败: "+err.Error())
		return
	}

	files := form.File["file"]
	if len(files) == 0 {
		ResponseError(ctx, "未找到上传的文件")
		return
	}

	file := files[0]
	fileHeader := &define.FileHeader{
		FileHeader: file,
		SaveFile: func(path string) error {
			return ctx.SaveUploadedFile(file, path)
		},
	}

	fileInfo, err := c.service.UploadAndParseOutlineFile(fileHeader) // 使用注入的服务实例
	if err != nil {
		ResponseError(ctx, err.Error())
		return
	}

	ResponseOK(ctx, fileInfo)
}
