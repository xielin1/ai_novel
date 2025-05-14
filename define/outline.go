package define

import "mime/multipart"

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

// OutlineFileInfo 大纲文件信息结构
type OutlineFileInfo struct {
	Content  string `json:"content"`
	Filename string `json:"filename"`
	Size     int64  `json:"size"`
	Link     string `json:"link"`
}

// FileHeader 封装文件上传处理
type FileHeader struct {
	*multipart.FileHeader
	SaveFile func(string) error
} 