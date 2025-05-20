package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/define"
	"gin-template/model"
	"gin-template/repository"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/google/uuid"
)

// OutlineService 定义了大纲管理的核心接口
type OutlineService interface {
	// ParseOutlineFile 解析上传的大纲文件内容
	ParseOutlineFile(filePath string, fileExt string) (string, error)

	// ValidateOutlineFile 验证大纲文件格式是否有效
	ValidateOutlineFile(filename string) (string, bool)

	// GetOutlineByProjectId 获取项目大纲
	GetOutlineByProjectId(projectId int) (interface{}, error)

	// SaveOutlineContent 保存大纲内容
	SaveOutlineContent(projectId int, content string) (*model.Outline, error)

	// GetVersionHistory 获取版本历史
	GetVersionHistory(projectId int, limit int) ([]*model.Version, error)

	// GenerateOutlineWithAI 使用AI生成大纲内容
	GenerateOutlineWithAI(userId uint, projectId int, content string, style string, wordLimit int) (map[string]interface{}, error)

	// ExportOutlineToFile 导出大纲到文件
	ExportOutlineToFile(projectId int, format string) (map[string]interface{}, error)

	// UploadAndParseOutlineFile 上传并解析大纲文件
	UploadAndParseOutlineFile(fileHeader *define.FileHeader) (*define.OutlineFileInfo, error)
}

// OutlineServiceImpl 是 OutlineService 的具体实现
type OutlineServiceImpl struct {
	tokenRepo   repository.TokenRepository
	reconRepo   repository.TokenReconciliationRepository
	outlineRepo repository.OutlineRepository
}

// NewOutlineService 创建一个新的 OutlineService 实例
func NewOutlineService(tokenRepo repository.TokenRepository, reconRepo repository.TokenReconciliationRepository, outlineRepo repository.OutlineRepository) OutlineService {
	logInfo("初始化OutlineService")
	return &OutlineServiceImpl{
		tokenRepo:   tokenRepo,
		reconRepo:   reconRepo,
		outlineRepo: outlineRepo,
	}
}

// ParseOutlineFile 解析上传的大纲文件内容
func (s *OutlineServiceImpl) ParseOutlineFile(filePath string, fileExt string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logError("文件不存在: %s", filePath)
		return "", fmt.Errorf("文件不存在: %s", filePath)
	}

	// 根据文件类型解析内容
	if fileExt == ".txt" {
		// 读取文本文件内容
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			logError("读取文件失败: %v", err)
			return "", fmt.Errorf("读取文件失败: %v", err)
		}
		logInfo("成功解析txt文件: %s", filePath)
		return string(data), nil
	} else if fileExt == ".docx" {
		// TODO: 实现docx文档解析
		logInfo("暂不支持docx格式解析")
		return "", fmt.Errorf("暂不支持docx格式解析")
	}

	logError("不支持的文件格式: %s", fileExt)
	return "", fmt.Errorf("不支持的文件格式: %s", fileExt)
}

// ValidateOutlineFile 验证大纲文件格式是否有效
func (s *OutlineServiceImpl) ValidateOutlineFile(filename string) (string, bool) {
	fileExt := filepath.Ext(filename)
	if fileExt != ".txt" && fileExt != ".docx" {
		logInfo("文件格式无效: %s", filename)
		return "", false
	}
	logInfo("文件格式有效: %s", filename)
	return fileExt, true
}

// GetOutlineByProjectId 获取项目大纲
func (s *OutlineServiceImpl) GetOutlineByProjectId(projectId int) (interface{}, error) {
	outline, err := s.outlineRepo.GetOutlineByProjectId(projectId)
	if err != nil {
		// 如果是新项目，可能没有大纲
		logInfo("项目 %d 没有大纲，返回空大纲", projectId)
		emptyOutline := map[string]interface{}{
			"id":              0,
			"project_id":      projectId,
			"content":         "",
			"current_version": 0,
			"created_at":      "",
			"updated_at":      "",
		}
		return emptyOutline, nil
	}

	logInfo("成功获取项目 %d 的大纲", projectId)
	return outline, nil
}

// SaveOutlineContent 保存大纲内容
func (s *OutlineServiceImpl) SaveOutlineContent(projectId int, content string) (*model.Outline, error) {
	// 保存大纲内容，创建新版本，非AI生成
	logInfo("保存项目 %d 的大纲内容", projectId)
	outline, err := s.outlineRepo.SaveOutline(projectId, content, false, "", 0, 0)
	if err != nil {
		logError("保存项目 %d 的大纲内容失败: %v", projectId, err)
		return nil, err
	}

	//// 更新项目的最后编辑时间
	//project, err := model.GetProjectById(projectId)
	//if err == nil {
	//	project.LastEditedAt = time.Now().Format("2006-01-02T15:04:05Z")
	//	project.Update()
	//	logInfo("更新项目 %d 的最后编辑时间", projectId)
	//}

	return outline, nil
}

// GetVersionHistory 获取版本历史
func (s *OutlineServiceImpl) GetVersionHistory(projectId int, limit int) ([]*model.Version, error) {
	if limit <= 0 {
		limit = 10
		logInfo("版本历史查询限制设置为默认值: %d", limit)
	}

	versions, err := s.outlineRepo.GetVersionHistory(projectId, limit)
	if err != nil {
		// 如果是新项目，可能没有版本历史
		logInfo("项目 %d 没有版本历史记录", projectId)
		return []*model.Version{}, nil
	}

	logInfo("成功获取项目 %d 的 %d 条版本历史记录", projectId, len(versions))
	return versions, nil
}

// GenerateOutlineWithAI 使用AI生成大纲内容
func (s *OutlineServiceImpl) GenerateOutlineWithAI(userId uint, projectId int, content string, style string, wordLimit int) (map[string]interface{}, error) {
	logInfo("开始为项目 %d 使用AI生成大纲内容", projectId)

	// 构造AI请求
	systemPrompt := "你是一个专业的内容创作助手，擅长根据提供的大纲进行续写和扩展。"
	if style != "" {
		systemPrompt += "请使用" + style + "的写作风格。"
	}

	userPrompt := "请根据以下大纲内容进行续写和扩展："
	if wordLimit > 0 {
		userPrompt += "续写内容大约" + strconv.Itoa(wordLimit) + "字。"
	}
	userPrompt += "\n\n" + content

	// 准备OpenAI请求
	openaiReq := define.GenerateAIPromptRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Model:        "gpt-3.5-turbo", // 可以从设置中获取或让用户选择
		MaxTokens:    2000,            // 根据字数限制调整
		Temperature:  0.7,             // 创意性参数
	}

	// 生成唯一交易ID，用于幂等性控制
	transactionUUID := uuid.New().String()

	// 调用AI服务进行续写
	openaiResp, err := GenerateAICompletion(openaiReq)
	if err != nil {
		logError("AI生成失败: %v", err)
		return nil, fmt.Errorf("AI生成失败: %v", err)
	}

	// 获取生成内容
	aiGeneratedContent := openaiResp.Content
	tokensUsed := openaiResp.TokensUsed

	// 保存大纲内容，创建新版本，标记为AI生成
	logInfo("保存AI生成的大纲内容，项目ID: %d, 使用Token: %d", projectId, tokensUsed)
	_, err = s.outlineRepo.SaveOutline(projectId, content+"\n\n"+aiGeneratedContent, true, style, wordLimit, tokensUsed)
	if err != nil {
		logError("保存AI生成的大纲内容失败: %v", err)
		return nil, err
	}

	//// 更新项目的最后编辑时间
	//project, err := model.GetProjectById(projectId)
	//if err == nil {
	//	project.LastEditedAt = time.Now().Format("2006-01-02T15:04:05Z")
	//	project.Update()
	//	logInfo("更新项目 %d 的最后编辑时间", projectId)
	//}

	// 扣除用户Token
	description := fmt.Sprintf("项目[%d]大纲AI续写消耗", projectId)
	projectIdStr := strconv.Itoa(projectId)

	// 使用TokenService扣减用户Token
	userToken, err := GetTokenService().DebitToken(
		userId,
		int64(tokensUsed),
		transactionUUID,
		"ai_generation_debit",
		description,
		"project",
		projectIdStr,
	)

	if err != nil {
		// 如果扣款失败但内容已生成，记录错误日志但仍然返回生成的内容
		// 在实际应用中，可能需要更复杂的错误处理策略
		logError("扣减用户Token失败: %v", err)
		return map[string]interface{}{
			"content":       aiGeneratedContent,
			"tokens_used":   tokensUsed,
			"token_balance": 0, // 余额获取失败
			"error":         "Token扣减失败，请联系客服",
		}, nil
	}

	// 获取用户最新Token余额
	tokenBalance := userToken.Balance

	logInfo("AI生成大纲内容成功，项目ID: %d, 使用Token: %d, 剩余Token: %d", projectId, tokensUsed, tokenBalance)
	return map[string]interface{}{
		"content":       aiGeneratedContent,
		"tokens_used":   tokensUsed,
		"token_balance": tokenBalance,
	}, nil
}

// ExportOutlineToFile 导出大纲到文件
func (s *OutlineServiceImpl) ExportOutlineToFile(projectId int, format string) (map[string]interface{}, error) {
	logInfo("开始导出项目 %d 的大纲到文件，格式: %s", projectId, format)

	// 检查格式
	if format != "txt" && format != "docx" && format != "pdf" {
		logError("不支持的导出格式: %s", format)
		return nil, fmt.Errorf("不支持的导出格式，仅支持txt、docx和pdf")
	}

	// 获取大纲内容
	outline, err := s.outlineRepo.GetOutlineByProjectId(projectId)
	if err != nil {
		logError("获取项目 %d 的大纲内容失败: %v", projectId, err)
		return nil, fmt.Errorf("获取大纲内容失败")
	}

	// 生成文件名
	fileName := fmt.Sprintf("outline_%d_%s.%s", projectId, time.Now().Format("20060102"), format)
	filePath := filepath.Join(common.UploadPath, fileName)

	// 写入文件
	if format == "txt" {
		err = ioutil.WriteFile(filePath, []byte(outline.Content), 0644)
		if err != nil {
			logError("生成文件失败: %v", err)
			return nil, fmt.Errorf("生成文件失败: %v", err)
		}
	} else {
		// TODO: 处理docx和pdf格式导出
		logInfo("暂不支持docx和pdf格式导出")
		return nil, fmt.Errorf("暂不支持docx和pdf格式导出")
	}

	// 返回文件路径
	fileUrl := "/upload/" + fileName
	fileSize := len(outline.Content)

	logInfo("成功导出项目 %d 的大纲到文件: %s", projectId, fileName)
	return map[string]interface{}{
		"file_url":  fileUrl,
		"file_size": fileSize,
	}, nil
}

// UploadAndParseOutlineFile 上传并解析大纲文件
func (s *OutlineServiceImpl) UploadAndParseOutlineFile(fileHeader *define.FileHeader) (*define.OutlineFileInfo, error) {
	logInfo("开始上传并解析大纲文件: %s", fileHeader.Filename)

	// 检查文件类型
	filename := filepath.Base(fileHeader.Filename)
	fileExt, valid := s.ValidateOutlineFile(filename)
	if !valid {
		logError("仅支持.txt和.docx格式的文件")
		return nil, fmt.Errorf("仅支持.txt和.docx格式的文件")
	}

	// 生成唯一文件名
	link := common.GetUUID() + fileExt
	savePath := filepath.Join(common.UploadPath, link)

	// 保存文件
	if err := fileHeader.SaveFile(savePath); err != nil {
		logError("保存文件失败: %v", err)
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}

	// 读取文件内容
	fileContent, err := s.ParseOutlineFile(savePath, fileExt)
	if err != nil {
		// 删除已保存的文件
		os.Remove(savePath)
		logError("解析文件失败，已删除保存的文件")
		return nil, err
	}

	// 构建文件信息响应
	fileInfo := &define.OutlineFileInfo{
		Content:  fileContent,
		Filename: filename,
		Size:     fileHeader.Size,
		Link:     link,
	}

	logInfo("成功上传并解析大纲文件: %s", filename)
	return fileInfo, nil
}
