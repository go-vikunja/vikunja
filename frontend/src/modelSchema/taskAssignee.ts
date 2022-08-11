import type {TypeOf} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'

export const TaskAssigneeSchema = AbstractSchema.extend({
	created: DateSchema.nullable(),
	userId: IdSchema.default(0), // IUser['id']
	taskId: IdSchema.default(0), // ITask['id']
}) 

export type TaskAssignee = TypeOf<typeof TaskAssigneeSchema> 
