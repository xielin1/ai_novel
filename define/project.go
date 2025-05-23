package define

type ProjectRequest struct {
	Title       string `json:"title" binding:"required"`
	Description string `json:"description"`
	Genre       string `json:"genre"`
}

// Pagination 分页信息
type Pagination struct {
	Total int64 `json:"total"`
	Page  int   `json:"page"`
	Limit int   `json:"limit"`
	Pages int64 `json:"pages"`
}

// ProjectListResponse 项目列表响应
type ProjectListResponse struct {
	Data       interface{} `json:"data"`
	Pagination Pagination  `json:"pagination"`
}

// 构建分页响应数据
func BuildPageResponse(data interface{}, total int64, page, limit int) ProjectListResponse {
	return ProjectListResponse{
		Data: data,
		Pagination: Pagination{
			Total: total,
			Page:  page,
			Limit: limit,
			Pages: (total + int64(limit) - 1) / int64(limit),
		},
	}
}
