package models

import "code.vikunja.io/api/pkg/web"

type TaskUnreadStatus struct {
	TaskID int64 `xorm:"bigint not null unique(task_user)"`
	UserID int64 `xorm:"bigint not null unique(task_user)"`
	web.CRUDable
	web.Permissions
}

func (TaskUnreadStatus) TableName() string {
	return "task_unread_statuses"
}
