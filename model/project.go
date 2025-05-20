package model

// Project 项目模型
type Project struct {
	Id           int    `json:"id"`
	Title        string `json:"title" gorm:"index"`
	Description  string `json:"description"`
	Genre        string `json:"genre"`
	UserId       int    `json:"user_id" gorm:"index"`
	Username     string `json:"username" gorm:"index"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
	LastEditedAt string `json:"last_edited_at"`
}
