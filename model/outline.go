package model

// Outline 大纲模型
type Outline struct {
	Id             int    `json:"id"`
	ProjectId      int    `json:"project_id" gorm:"index"`
	Content        string `json:"content" gorm:"type:text"`
	CurrentVersion int    `json:"current_version"`
	CreatedAt      string `json:"created_at"`
	UpdatedAt      string `json:"updated_at"`
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
