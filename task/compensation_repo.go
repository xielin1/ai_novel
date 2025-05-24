package task

import (
	"gin-template/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

type DBCompensation struct {
	db *gorm.DB
}

func NewDBCompensation(db *gorm.DB) (*DBCompensation, error) {
	return &DBCompensation{db: db}, nil
}

func (d *DBCompensation) AddTask(task model.CompensationTask) error {
	task.CreatedAt = time.Now()
	if task.NextExecute.IsZero() {
		task.NextExecute = time.Now()
	}
	return d.db.Create(&task).Error
}

func (d *DBCompensation) GetPendingTasks() []model.CompensationTask {
	var tasks []model.CompensationTask
	now := time.Now()

	// 使用事务保证数据一致性
	d.db.Transaction(func(tx *gorm.DB) error {
		// 锁定选中的行防止重复处理
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("status = ? AND next_execute <= ?", model.Pending, now).
			Find(&tasks).Error; err != nil {
			return err
		}

		// 批量更新状态为processing
		var ids []string
		for _, task := range tasks {
			ids = append(ids, task.TaskID)
		}
		if len(ids) > 0 {
			return tx.Model(&model.CompensationTask{}).
				Where("task_id IN ?", ids).
				Updates(map[string]interface{}{
					"status":      model.Processing,
					"retry_count": gorm.Expr("retry_count + 1"),
				}).Error
		}
		return nil
	})
	return tasks
}

func (d *DBCompensation) UpdateTask(task model.CompensationTask) error {
	return d.db.Save(&task).Error
}
