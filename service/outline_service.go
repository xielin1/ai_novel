package service

import (
	"fmt"
	"gin-template/common"
	"gin-template/define"
	"gin-template/model"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// ParseOutlineFile 解析上传的大纲文件内容
func ParseOutlineFile(filePath string, fileExt string) (string, error) {
	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return "", fmt.Errorf("文件不存在: %s", filePath)
	}

	// 根据文件类型解析内容
	if fileExt == ".txt" {
		// 读取文本文件内容
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("读取文件失败: %v", err)
		}
		return string(data), nil
	} else if fileExt == ".docx" {
		// TODO: 实现docx文档解析
		return "", fmt.Errorf("暂不支持docx格式解析")
	}

	return "", fmt.Errorf("不支持的文件格式: %s", fileExt)
}

// ValidateOutlineFile 验证大纲文件格式是否有效
func ValidateOutlineFile(filename string) (string, bool) {
	fileExt := filepath.Ext(filename)
	if fileExt != ".txt" && fileExt != ".docx" {
		return "", false
	}
	return fileExt, true
}

// GetOutlineByProjectId 获取项目大纲
func GetOutlineByProjectId(projectId int) (interface{}, error) {
	outline, err := model.GetOutlineByProjectId(projectId)
	if err != nil {
		// 如果是新项目，可能没有大纲
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
	
	return outline, nil
}

// SaveOutlineContent 保存大纲内容
func SaveOutlineContent(projectId int, content string) (*model.Outline, error) {
	// 保存大纲内容，创建新版本，非AI生成
	outline, err := model.SaveOutline(projectId, content, false, "", 0, 0)
	if err != nil {
		return nil, err
	}
	
	// 更新项目的最后编辑时间
	project, err := model.GetProjectById(projectId)
	if err == nil {
		project.LastEditedAt = time.Now().Format("2006-01-02T15:04:05Z")
		project.Update()
	}
	
	return outline, nil
}

// GetVersionHistory 获取版本历史
func GetVersionHistory(projectId int, limit int) ([]*model.Version, error) {
	if limit <= 0 {
		limit = 10
	}
	
	versions, err := model.GetVersionHistory(projectId, limit)
	if err != nil {
		// 如果是新项目，可能没有版本历史
		return []*model.Version{}, nil
	}
	
	return versions, nil
}

// GenerateOutlineWithAI 使用AI生成大纲内容
func GenerateOutlineWithAI(projectId int, content string, style string, wordLimit int) (map[string]interface{}, error) {
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
	
	// 调用AI服务进行续写
	openaiResp, err := GenerateAICompletion(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("AI生成失败: %v", err)
	}
	
	// 获取生成内容
	aiGeneratedContent := openaiResp.Content
	tokensUsed := openaiResp.TokensUsed
	
	// 保存大纲内容，创建新版本，标记为AI生成
	_, err = model.SaveOutline(projectId, content+"\n\n"+aiGeneratedContent, true, style, wordLimit, tokensUsed)
	if err != nil {
		return nil, err
	}
	
	// 更新项目的最后编辑时间
	project, err := model.GetProjectById(projectId)
	if err == nil {
		project.LastEditedAt = time.Now().Format("2006-01-02T15:04:05Z")
		project.Update()
	}
	
	// 获取用户Token余额
	tokenBalance := 1000 // 假设初始余额为1000，实际应从用户账户中获取
	tokenBalance -= tokensUsed
	
	return map[string]interface{}{
		"content":       aiGeneratedContent,
		"tokens_used":   tokensUsed,
		"token_balance": tokenBalance,
	}, nil
}

// ExportOutlineToFile 导出大纲到文件
func ExportOutlineToFile(projectId int, format string) (map[string]interface{}, error) {
	// 检查格式
	if format != "txt" && format != "docx" && format != "pdf" {
		return nil, fmt.Errorf("不支持的导出格式，仅支持txt、docx和pdf")
	}
	
	// 获取大纲内容
	outline, err := model.GetOutlineByProjectId(projectId)
	if err != nil {
		return nil, fmt.Errorf("获取大纲内容失败")
	}
	
	// 生成文件名
	fileName := fmt.Sprintf("outline_%d_%s.%s", projectId, time.Now().Format("20060102"), format)
	filePath := filepath.Join(common.UploadPath, fileName)
	
	// 写入文件
	if format == "txt" {
		err = ioutil.WriteFile(filePath, []byte(outline.Content), 0644)
		if err != nil {
			return nil, fmt.Errorf("生成文件失败: %v", err)
		}
	} else {
		// TODO: 处理docx和pdf格式导出
		return nil, fmt.Errorf("暂不支持docx和pdf格式导出")
	}
	
	// 返回文件路径
	fileUrl := "/upload/" + fileName
	fileSize := len(outline.Content)
	
	return map[string]interface{}{
		"file_url":  fileUrl,
		"file_size": fileSize,
	}, nil
}

// UploadAndParseOutlineFile 上传并解析大纲文件
func UploadAndParseOutlineFile(fileHeader *define.FileHeader) (*define.OutlineFileInfo, error) {
	// 检查文件类型
	filename := filepath.Base(fileHeader.Filename)
	fileExt, valid := ValidateOutlineFile(filename)
	if !valid {
		return nil, fmt.Errorf("仅支持.txt和.docx格式的文件")
	}
	
	// 生成唯一文件名
	link := common.GetUUID() + fileExt
	savePath := filepath.Join(common.UploadPath, link)
	
	// 保存文件
	if err := fileHeader.SaveFile(savePath); err != nil {
		return nil, fmt.Errorf("保存文件失败: %v", err)
	}
	
	// 读取文件内容
	fileContent, err := ParseOutlineFile(savePath, fileExt)
	if err != nil {
		// 删除已保存的文件
		os.Remove(savePath)
		return nil, err
	}
	
	// 构建文件信息响应
	fileInfo := &define.OutlineFileInfo{
		Content:  fileContent,
		Filename: filename,
		Size:     fileHeader.Size,
		Link:     link,
	}
	
	return fileInfo, nil
} 