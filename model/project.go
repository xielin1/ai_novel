package model

// Project 项目模型
type Project struct {
	Id          int    `json:"id"`
	Title       string `json:"title" gorm:"index"`
	Description string `json:"description"`
	Genre       string `json:"genre"`
	UserId      int    `json:"user_id" gorm:"index"`
	Username    string `json:"username" gorm:"index"`
	CreatedAt   string `json:"created_at"`
	UpdatedAt   string `json:"updated_at"`
	LastEditedAt string `json:"last_edited_at"`
}

// GetUserProjects 获取用户的项目列表
func GetUserProjects(userId int, startIdx int, num int) ([]*Project, int64, error) {
	var projects []*Project
	var total int64
	
	// 获取项目总数
	err := DB.Model(&Project{}).Where("user_id = ?", userId).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	
	// 获取分页项目列表
	err = DB.Where("user_id = ?", userId).Order("id desc").Limit(num).Offset(startIdx).Find(&projects).Error
	return projects, total, err
}

// GetProjectById 根据ID获取项目
func GetProjectById(id int) (*Project, error) {
	var project Project
	err := DB.Where("id = ?", id).First(&project).Error
	return &project, err
}

// Insert 插入新项目
func (project *Project) Insert() error {
	return DB.Create(project).Error
}

// Update 更新项目
func (project *Project) Update() error {
	return DB.Save(project).Error
}

// Delete 删除项目
func (project *Project) Delete() error {
	return DB.Delete(project).Error
} 