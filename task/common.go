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

// 增强的补偿调度器
type DBScheduler struct {
	table    *DBCompensation
	handlers map[string]func(model.CompensationTask) error
	interval time.Duration
}

func NewDBScheduler(table *DBCompensation, interval time.Duration) *DBScheduler {
	return &DBScheduler{
		table:    table,
		handlers: make(map[string]func(model.CompensationTask) error),
		interval: interval,
	}
}

// 注册处理器
func (d *DBScheduler) RegisterHandler(taskType string, handler func(model.CompensationTask) error) {
	d.handlers[taskType] = handler
}

// 启动调度器（增加恢复机制）
func (d *DBScheduler) Start() {
	// 先处理卡在processing状态的任务
	d.recoverStuckTasks()

	ticker := time.NewTicker(d.interval)
	go func() {
		for range ticker.C {
			d.processTasks()
		}
	}()
}

// 处理残留的processing任务
func (d *DBScheduler) recoverStuckTasks() {
	var stuckTasks []model.CompensationTask
	d.table.db.Where("status = ? AND next_execute <= ?", model.Processing, time.Now().Add(-1*time.Hour)).
		Find(&stuckTasks)

	for _, task := range stuckTasks {
		task.Status = model.Pending
		d.table.UpdateTask(task)
	}
}

// 处理任务（优化批量处理）
func (d *DBScheduler) processTasks() {
	tasks := d.table.GetPendingTasks()

	for _, task := range tasks {
		handler, exists := d.handlers[task.TaskType]
		if !exists {
			task.LastError = "no handler registered"
			task.Status = model.Failed
			d.table.UpdateTask(task)
			continue
		}

		err := handler(task)
		updateData := map[string]interface{}{
			"last_error": task.LastError,
			"status":     task.Status,
		}

		if err != nil {
			task.LastError = err.Error()
			if task.RetryCount >= task.MaxRetries {
				task.Status = model.Failed
			} else {
				task.Status = model.Pending
				delay := time.Duration(task.RetryCount*task.RetryCount) * time.Second
				updateData["next_execute"] = time.Now().Add(delay)
			}
		} else {
			task.Status = model.Succeeded
		}

		d.table.db.Model(&model.CompensationTask{}).
			Where("task_id = ?", task.TaskID).
			Updates(updateData)
	}
}
