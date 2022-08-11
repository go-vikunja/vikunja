import type {TypeOf} from 'zod'
import {number, array, boolean} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'
import {TextFieldSchema} from './common/textField'

import {AbstractSchema} from './abstract'
import {UserSchema} from './user'
import {TaskSchema} from './task'

export const BucketSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	title: TextFieldSchema,
	listId: IdSchema.default(0),
	limit: number().default(0),
	tasks: array(TaskSchema).default([]),
	isDoneBucket: boolean().default(false),
	position: number().default(0),
	
	createdBy: UserSchema.nullable(),
	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
})

export type IBucket = TypeOf<typeof BucketSchema> 