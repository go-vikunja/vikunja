package models

type TaskUnreadStatuses struct {
	TaskID int64 `xorm:"bigint not null unique(task_user)"`
	UserID int64 `xorm:"bigint not null unique(task_user)"`
}

func (TaskUnreadStatuses) TableName() string {
	return "task_unread_statuses"
}
