package models

type TaskUnreadStatus struct {
	TaskID int64 `xorm:"bigint not null unique(task_user)"`
	UserID int64 `xorm:"bigint not null unique(task_user)"`
}

func (TaskUnreadStatus) TableName() string {
	return "task_unread_statuses"
}
