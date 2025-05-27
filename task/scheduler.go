package task

import (
	"fmt"
	"gin-template/common"
	"gin-template/model"
	"time"
)

type DBScheduler struct {
	table    *DBCompensation
	handlers map[string]func(model.CompensationTask) error
	interval time.Duration
}

func NewDBScheduler(table *DBCompensation, interval time.Duration) *DBScheduler {
	common.SysLog(fmt.Sprintf("初始化任务调度器，扫描间隔：%v", interval))
	return &DBScheduler{
		table:    table,
		handlers: make(map[string]func(model.CompensationTask) error),
		interval: interval,
	}
}

func (d *DBScheduler) RegisterHandler(taskType string, handler func(model.CompensationTask) error) {
	d.handlers[taskType] = handler
	common.SysLog(fmt.Sprintf("注册任务处理器，类型：%s", taskType))
}

func (d *DBScheduler) Start() {
	common.SysLog("任务调度器启动")
	defer common.SysLog("任务调度器停止")

	// 处理残留任务并记录日志
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
	start := time.Now()
	var stuckTasks []model.CompensationTask
	result := d.table.db.Where("status = ? AND next_execute <= ?", model.Processing, time.Now().Add(-1*time.Hour)).
		Find(&stuckTasks)

	if result.Error != nil {
		common.SysError(fmt.Sprintf("查询残留任务失败：%v", result.Error))
		return
	}

	if len(stuckTasks) > 0 {
		common.SysLog(fmt.Sprintf("发现 %d 个残留任务，开始重置状态", len(stuckTasks)))
		for _, task := range stuckTasks {
			task.Status = model.Pending
			if err := d.table.UpdateTask(task); err != nil {
				common.SysError(fmt.Sprintf("重置任务 %s 状态失败：%v", task.TaskID, err))
			} else {
				common.SysLog(fmt.Sprintf("重置任务 %s 状态为 pending", task.TaskID))
			}
		}
		common.SysLog(fmt.Sprintf("残留任务处理完成，耗时：%v", time.Since(start)))
	} else {
		common.SysLog("未发现残留任务")
	}
}

// 处理任务（优化批量处理）
func (d *DBScheduler) processTasks() {
	start := time.Now()
	tasks := d.table.GetPendingTasks()
	common.SysLog(fmt.Sprintf("开始处理任务，待处理数量：%d", len(tasks)))

	for _, task := range tasks {
		handler, exists := d.handlers[task.TaskType]
		if !exists {
			task.LastError = "no handler registered"
			task.Status = model.Failed
			if err := d.table.UpdateTask(task); err != nil {
				common.SysError(fmt.Sprintf("更新任务 %s 失败：%v", task.TaskID, err))
			} else {
				common.SysLog(fmt.Sprintf("任务 %s 无处理器，标记为 failed", task.TaskID))
			}
			continue
		}

		common.SysLog(fmt.Sprintf("开始执行任务 %s，类型：%s", task.TaskID, task.TaskType))
		err := handler(task)
		updateData := map[string]interface{}{
			"last_error": task.LastError,
			"status":     task.Status,
		}

		if err != nil {
			task.LastError = err.Error()
			common.SysError(fmt.Sprintf("任务 %s 执行失败：%v", task.TaskID, err))
			if task.RetryCount >= task.MaxRetries {
				updateData["status"] = model.Failed
				common.SysLog(fmt.Sprintf("任务 %s 重试次数用尽，标记为 failed", task.TaskID))
			} else {
				task.Status = model.Pending
				delay := time.Duration(task.RetryCount*task.RetryCount) * time.Second
				updateData["next_execute"] = time.Now().Add(delay)
				updateData["retry_count"] = task.RetryCount + 1
				common.SysLog(fmt.Sprintf("任务 %s 重试，下次执行时间：%v，延迟：%v", task.TaskID, updateData["next_execute"], delay))
			}
		} else {
			updateData["status"] = model.Succeeded
			common.SysLog(fmt.Sprintf("任务 %s 执行成功", task.TaskID))
		}

		if err := d.table.db.Model(&model.CompensationTask{}).
			Where("task_id = ?", task.TaskID).
			Updates(updateData).Error; err != nil {
			common.SysError(fmt.Sprintf("更新任务 %s 状态失败：%v", task.TaskID, err))
		}
	}

	common.SysLog(fmt.Sprintf("任务处理完成，耗时：%v，处理数量：%d", time.Since(start), len(tasks)))
}
