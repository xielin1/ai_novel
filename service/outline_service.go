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

type OutlineService struct {
	tokenRepo   *repository.TokenRepository
	reconRepo   *repository.TokenReconciliationRepository
	outlineRepo *repository.OutlineRepository
}

func NewOutlineService(tokenRepo *repository.TokenRepository, reconRepo *repository.TokenReconciliationRepository, outlineRepo *repository.OutlineRepository) *OutlineService {
	common.SysLog("[OutlineService] Initializing OutlineService")
	return &OutlineService{
		tokenRepo:   tokenRepo,
		reconRepo:   reconRepo,
		outlineRepo: outlineRepo,
	}
}

// ParseOutlineFile parses the uploaded outline file content
func (s *OutlineService) ParseOutlineFile(filePath string, fileExt string) (string, error) {
	// Check if file exists
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		logMsg := fmt.Sprintf("[OutlineService] File not found: %s", filePath)
		common.SysError(logMsg)
		return "", fmt.Errorf(logMsg)
	}

	// Parse content based on file type
	if fileExt == ".txt" {
		// Read text file content
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			logMsg := fmt.Sprintf("[OutlineService] Failed to read file: %v", err)
			common.SysError(logMsg)
			return "", fmt.Errorf(logMsg)
		}
		logMsg := fmt.Sprintf("[OutlineService] Successfully parsed txt file: %s", filePath)
		common.SysLog(logMsg)
		return string(data), nil
	} else if fileExt == ".docx" {
		// TODO: Implement docx parsing
		logMsg := "[OutlineService] Docx format parsing is not supported yet"
		common.SysLog(logMsg)
		return "", fmt.Errorf(logMsg)
	}

	logMsg := fmt.Sprintf("[OutlineService] Unsupported file format: %s", fileExt)
	common.SysError(logMsg)
	return "", fmt.Errorf(logMsg)
}

// ValidateOutlineFile validates if the outline file format is valid
func (s *OutlineService) ValidateOutlineFile(filename string) (string, bool) {
	fileExt := filepath.Ext(filename)
	if fileExt != ".txt" && fileExt != ".docx" {
		logMsg := fmt.Sprintf("[OutlineService] Invalid file format: %s", filename)
		common.SysLog(logMsg)
		return "", false
	}
	logMsg := fmt.Sprintf("[OutlineService] Valid file format: %s", filename)
	common.SysLog(logMsg)
	return fileExt, true
}

// GetOutlineByProjectId retrieves the outline by project ID
func (s *OutlineService) GetOutlineByProjectId(projectId int64) (interface{}, error) {
	outline, err := s.outlineRepo.GetOutlineByProjectId(projectId)
	if err != nil {
		// Empty outline for new projects
		logMsg := fmt.Sprintf("[OutlineService] No outline found for project %d, returning empty outline", projectId)
		common.SysLog(logMsg)
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

	logMsg := fmt.Sprintf("[OutlineService] Successfully retrieved outline for project %d", projectId)
	common.SysLog(logMsg)
	return outline, nil
}

// SaveOutlineContent saves the outline content
func (s *OutlineService) SaveOutlineContent(projectId int64, content string) (*model.Outline, error) {
	// Save outline content, create new version, not AI-generated
	logMsg := fmt.Sprintf("[OutlineService] Saving outline content for project %d", projectId)
	common.SysLog(logMsg)
	outline, err := s.outlineRepo.SaveOutline(projectId, content, false, "", 0, 0)
	if err != nil {
		logMsg := fmt.Sprintf("[OutlineService] Failed to save outline content for project %d: %v", projectId, err)
		common.SysError(logMsg)
		return nil, err
	}

	//// Update project's last edit time
	//project, err := model.GetProjectById(projectId)
	//if err == nil {
	//	project.LastEditedAt = time.Now().Format("2006-01-02T15:04:05Z")
	//	project.Update()
	//	logMsg := fmt.Sprintf("[OutlineService] Updated last edit time for project %d", projectId)
	//	common.SysLog(logMsg)
	//}

	return outline, nil
}

// GetVersionHistory retrieves version history
func (s *OutlineService) GetVersionHistory(projectId int64, limit int) ([]*model.Version, error) {
	if limit <= 0 {
		limit = 10
		logMsg := fmt.Sprintf("[OutlineService] Version history query limit set to default: %d", limit)
		common.SysLog(logMsg)
	}

	versions, err := s.outlineRepo.GetVersionHistory(projectId, limit)
	if err != nil {
		// No version history for new projects
		logMsg := fmt.Sprintf("[OutlineService] No version history found for project %d", projectId)
		common.SysLog(logMsg)
		return []*model.Version{}, nil
	}

	logMsg := fmt.Sprintf("[OutlineService] Successfully retrieved %d version history records for project %d", len(versions), projectId)
	common.SysLog(logMsg)
	return versions, nil
}

// GenerateOutlineWithAI generates outline content using AI
func (s *OutlineService) GenerateOutlineWithAI(userId int64, projectId int64, content string, style string, wordLimit int) (map[string]interface{}, error) {
	logMsg := fmt.Sprintf("[OutlineService] Starting AI outline generation for project %d", projectId)
	common.SysLog(logMsg)

	// Construct AI request
	systemPrompt := "You are a professional content creation assistant skilled at continuing and expanding on provided outlines."
	if style != "" {
		systemPrompt += fmt.Sprintf(" Please use a %s writing style.", style)
	}

	userPrompt := "Please continue and expand on the following outline content:"
	if wordLimit > 0 {
		userPrompt += fmt.Sprintf(" The continuation should be approximately %d words.", wordLimit)
	}
	userPrompt += "\n\n" + content

	// Prepare OpenAI request
	openaiReq := define.GenerateAIPromptRequest{
		SystemPrompt: systemPrompt,
		UserPrompt:   userPrompt,
		Model:        "gpt-3.5-turbo", // Can be obtained from settings or user selection
		MaxTokens:    2000,            // Adjust based on word limit
		Temperature:  0.7,             // Creativity parameter
	}

	// Generate unique transaction ID for idempotency control
	transactionUUID := uuid.New().String()

	// Call AI service for continuation
	openaiResp, err := GenerateAICompletion(openaiReq)
	if err != nil {
		logMsg := fmt.Sprintf("[OutlineService] AI generation failed: %v", err)
		common.SysError(logMsg)
		return nil, fmt.Errorf(logMsg)
	}

	// Get generated content
	aiGeneratedContent := openaiResp.Content
	tokensUsed := openaiResp.TokensUsed

	// Save outline content, create new version, mark as AI-generated
	logMsg = fmt.Sprintf("[OutlineService] Saving AI-generated outline content, Project ID: %d, Tokens used: %d", projectId, tokensUsed)
	common.SysLog(logMsg)
	_, err = s.outlineRepo.SaveOutline(projectId, content+"\n\n"+aiGeneratedContent, true, style, wordLimit, tokensUsed)
	if err != nil {
		logMsg = fmt.Sprintf("[OutlineService] Failed to save AI-generated outline content: %v", err)
		common.SysError(logMsg)
		return nil, err
	}

	//// Update project's last edit time
	//project, err := model.GetProjectById(projectId)
	//if err == nil {
	//	project.LastEditedAt = time.Now().Format("2006-01-02T15:04:05Z")
	//	project.Update()
	//	logMsg := fmt.Sprintf("[OutlineService] Updated last edit time for project %d", projectId)
	//	common.SysLog(logMsg)
	//}

	// Deduct user tokens
	description := fmt.Sprintf("AI outline continuation for project [%d]", projectId)

	// Use TokenService to deduct user tokens
	service := GetTokenService()
	userToken, err := service.DebitToken(
		userId,
		int64(tokensUsed),
		transactionUUID,
		"ai_generation_debit",
		description,
		"project",
		strconv.FormatInt(projectId, 10),
	)

	if err != nil {
		// Log error but return generated content if deduction fails (needs actual error handling strategy)
		logMsg = fmt.Sprintf("[OutlineService] Failed to deduct user tokens: %v", err)
		common.SysError(logMsg)
		return map[string]interface{}{
			"content":       aiGeneratedContent,
			"tokens_used":   tokensUsed,
			"token_balance": 0, // Failed to get balance
			"error":         "Token deduction failed, please contact support",
		}, nil
	}

	// Get user's latest token balance
	tokenBalance := userToken.Balance

	logMsg = fmt.Sprintf("[OutlineService] AI outline generation successful, Project ID: %d, Tokens used: %d, Remaining tokens: %d", projectId, tokensUsed, tokenBalance)
	common.SysLog(logMsg)
	return map[string]interface{}{
		"content":       aiGeneratedContent,
		"tokens_used":   tokensUsed,
		"token_balance": tokenBalance,
	}, nil
}

// ExportOutlineToFile exports the outline to a file
func (s *OutlineService) ExportOutlineToFile(projectId int64, format string) (map[string]interface{}, error) {
	logMsg := fmt.Sprintf("[OutlineService] Starting to export outline for project %d to file, format: %s", projectId, format)
	common.SysLog(logMsg)

	// Check format
	if format != "txt" && format != "docx" && format != "pdf" {
		logMsg := fmt.Sprintf("[OutlineService] Unsupported export format: %s", format)
		common.SysError(logMsg)
		return nil, fmt.Errorf("Unsupported export format, only txt, docx, and pdf are allowed")
	}

	// Get outline content
	outline, err := s.outlineRepo.GetOutlineByProjectId(projectId)
	if err != nil {
		logMsg := fmt.Sprintf("[OutlineService] Failed to get outline content for project %d: %v", projectId, err)
		common.SysError(logMsg)
		return nil, fmt.Errorf("Failed to get outline content")
	}

	// Generate file name
	fileName := fmt.Sprintf("outline_%d_%s.%s", projectId, time.Now().Format("20060102"), format)
	filePath := filepath.Join(common.UploadPath, fileName)

	// Write file
	if format == "txt" {
		err = ioutil.WriteFile(filePath, []byte(outline.Content), 0644)
		if err != nil {
			logMsg := fmt.Sprintf("[OutlineService] Failed to generate file: %v", err)
			common.SysError(logMsg)
			return nil, fmt.Errorf(logMsg)
		}
	} else {
		// TODO: Handle docx and pdf exports
		logMsg := "[OutlineService] Docx and pdf format exports are not supported yet"
		common.SysLog(logMsg)
		return nil, fmt.Errorf(logMsg)
	}

	// Return file path
	fileUrl := "/upload/" + fileName
	fileSize := len(outline.Content)

	logMsg = fmt.Sprintf("[OutlineService] Successfully exported outline for project %d to file: %s", projectId, fileName)
	common.SysLog(logMsg)
	return map[string]interface{}{
		"file_url":  fileUrl,
		"file_size": fileSize,
	}, nil
}

// UploadAndParseOutlineFile uploads and parses an outline file
func (s *OutlineService) UploadAndParseOutlineFile(fileHeader *define.FileHeader) (*define.OutlineFileInfo, error) {
	logMsg := fmt.Sprintf("[OutlineService] Starting to upload and parse outline file: %s", fileHeader.Filename)
	common.SysLog(logMsg)

	// Check file type
	filename := filepath.Base(fileHeader.Filename)
	fileExt, valid := s.ValidateOutlineFile(filename)
	if !valid {
		logMsg := "[OutlineService] Only .txt and .docx files are supported"
		common.SysError(logMsg)
		return nil, fmt.Errorf(logMsg)
	}

	// Generate unique file name
	link := common.GetUUID() + fileExt
	savePath := filepath.Join(common.UploadPath, link)

	// Save file
	if err := fileHeader.SaveFile(savePath); err != nil {
		logMsg := fmt.Sprintf("[OutlineService] Failed to save file: %v", err)
		common.SysError(logMsg)
		return nil, fmt.Errorf(logMsg)
	}

	// Read file content
	fileContent, err := s.ParseOutlineFile(savePath, fileExt)
	if err != nil {
		// Delete saved file
		os.Remove(savePath)
		logMsg := "[OutlineService] Failed to parse file, saved file has been deleted"
		common.SysError(logMsg)
		return nil, err
	}

	// Build file info response
	fileInfo := &define.OutlineFileInfo{
		Content:  fileContent,
		Filename: filename,
		Size:     fileHeader.Size,
		Link:     link,
	}

	logMsg = fmt.Sprintf("[OutlineService] Successfully uploaded and parsed outline file: %s", filename)
	common.SysLog(logMsg)
	return fileInfo, nil
}
