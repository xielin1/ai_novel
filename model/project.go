package model

// Project 项目模型
type Project struct {
	Id           int    `json:"id"`
	Title        string `json:"title" gorm:"index"`
	Description  string `json:"description"`
	Genre        string `json:"genre"`
	UserId       int    `json:"user_id" gorm:"index"`
	Username     string `json:"username" gorm:"index"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
	LastEditedAt int64  `json:"last_edited_at"`
}
