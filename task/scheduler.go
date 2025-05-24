package task

import (
	"gin-template/model"
	"time"
)

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

func (d *DBScheduler) RegisterHandler(taskType string, handler func(model.CompensationTask) error) {
	d.handlers[taskType] = handler
}

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
