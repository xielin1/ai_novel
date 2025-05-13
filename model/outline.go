package model

import (
	"time"
)

// Outline 大纲模型
type Outline struct {
	Id            int    `json:"id"`
	ProjectId     int    `json:"project_id" gorm:"index"`
	Content       string `json:"content" gorm:"type:text"`
	CurrentVersion int    `json:"current_version"`
	CreatedAt     string `json:"created_at"`
	UpdatedAt     string `json:"updated_at"`
}

// Version 版本历史模型
type Version struct {
	Id            int    `json:"id"`
	OutlineId     int    `json:"outline_id" gorm:"index"`
	VersionNumber int    `json:"version_number"`
	Content       string `json:"content" gorm:"type:text"`
	IsAiGenerated bool   `json:"is_ai_generated"`
	AiStyle       string `json:"ai_style"`
	WordLimit     int    `json:"word_limit"`
	TokensUsed    int    `json:"tokens_used"`
	CreatedAt     string `json:"created_at"`
}

// GetOutlineByProjectId 根据项目ID获取大纲
func GetOutlineByProjectId(projectId int) (*Outline, error) {
	var outline Outline
	err := DB.Where("project_id = ?", projectId).First(&outline).Error
	return &outline, err
}

// SaveOutline 保存大纲内容并创建新版本
func SaveOutline(projectId int, content string, isAiGenerated bool, aiStyle string, wordLimit int, tokensUsed int) (*Outline, error) {
	// 获取当前时间
	currentTime := time.Now().Format("2006-01-02T15:04:05Z")
	
	// 查找是否存在大纲
	var outline Outline
	result := DB.Where("project_id = ?", projectId).First(&outline)
	
	if result.Error != nil {
		// 如果不存在，创建新大纲
		outline = Outline{
			ProjectId:     projectId,
			Content:       content,
			CurrentVersion: 1,
			CreatedAt:     currentTime,
			UpdatedAt:     currentTime,
		}
		
		if err := DB.Create(&outline).Error; err != nil {
			return nil, err
		}
	} else {
		// 如果存在，更新大纲内容和版本
		outline.Content = content
		outline.CurrentVersion += 1
		outline.UpdatedAt = currentTime
		
		if err := DB.Save(&outline).Error; err != nil {
			return nil, err
		}
	}
	
	// 创建新版本记录
	version := Version{
		OutlineId:     outline.Id,
		VersionNumber: outline.CurrentVersion,
		Content:       content,
		IsAiGenerated: isAiGenerated,
		AiStyle:       aiStyle,
		WordLimit:     wordLimit,
		TokensUsed:    tokensUsed,
		CreatedAt:     currentTime,
	}
	
	if err := DB.Create(&version).Error; err != nil {
		return nil, err
	}
	
	return &outline, nil
}

// GetVersionHistory 获取版本历史
func GetVersionHistory(projectId int, limit int) ([]*Version, error) {
	var outline Outline
	var versions []*Version
	
	// 先找到对应的大纲
	err := DB.Where("project_id = ?", projectId).First(&outline).Error
	if err != nil {
		return versions, err
	}
	
	// 获取版本历史，按版本号降序排列
	err = DB.Where("outline_id = ?", outline.Id).
		Order("version_number desc").
		Limit(limit).
		Find(&versions).Error
		
	return versions, err
} 