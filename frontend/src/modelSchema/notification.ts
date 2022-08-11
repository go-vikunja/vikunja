import type {TypeOf} from 'zod'
import {union, boolean, object, string} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'
import {TaskSchema} from './task'
import {TaskCommentSchema} from './taskComment'
import {TeamSchema} from './team'
import {UserSchema} from './user'

const NotificationTypeSchema = object({
	doer: UserSchema,
})

const NotificationTypeTask = NotificationTypeSchema.extend({
	task: TaskSchema,
	comment: TaskCommentSchema,
})

const NotificationTypeAssigned = NotificationTypeSchema.extend({
	task: TaskSchema,
	assignee: UserSchema,
})

const NotificationTypeDeleted = NotificationTypeSchema.extend({
	task: TaskSchema,
})

const NotificationTypeCreated = NotificationTypeSchema.extend({
	task: TaskSchema,
})

const NotificationTypeMemberAdded = NotificationTypeSchema.extend({
	member: UserSchema,
	team: TeamSchema,
})

export const NotificationSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	name: string().default(''),
	notification: union([
		NotificationTypeTask,
		NotificationTypeAssigned,
		NotificationTypeDeleted,
		NotificationTypeCreated,
		NotificationTypeMemberAdded,
	]),
	read: boolean().default(false),
	readAt: DateSchema.nullable(),
}) 

export type Notification = TypeOf<typeof NotificationSchema> 
