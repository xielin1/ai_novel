package model

import "time"

type TaskStatus string

const (
	Pending    TaskStatus = "pending"
	Processing TaskStatus = "processing"
	Succeeded  TaskStatus = "succeeded"
	Failed     TaskStatus = "failed"
)

type CompensationTask struct {
	TaskID      string     `gorm:"primaryKey;size:36"`
	TaskType    string     `gorm:"index;size:50;not null"`
	Payload     string     `gorm:"type:text;not null"`
	Status      TaskStatus `gorm:"type:enum('pending','processing','succeeded','failed');default:'pending'"`
	RetryCount  int        `gorm:"default:0"`
	MaxRetries  int        `gorm:"default:3"`
	CreatedAt   time.Time  `gorm:"index"`
	NextExecute time.Time  `gorm:"index"`
	LastError   string     `gorm:"type:text"`
}
