import type {TypeOf} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'
import {UserSchema} from './user'
import {FileSchema} from './file'

export const AttachmentSchema = AbstractSchema.extend({
	id: IdSchema.default(0),

	taskId: IdSchema.default(0), // iTaskSchema.shape.id
	createdBy: UserSchema,
	file: FileSchema,
	created: DateSchema.nullable(),
})

export type IAttachment = TypeOf<typeof AttachmentSchema>