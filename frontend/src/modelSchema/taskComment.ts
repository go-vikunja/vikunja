import type {TypeOf} from 'zod'
import {string} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'
import {UserSchema} from './user'

export const TaskCommentSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	taskId: IdSchema.default(0),
	comment: string().default(''),
	author: UserSchema,
	
	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
}) 

export type TaskComment = TypeOf<typeof TaskCommentSchema> 
