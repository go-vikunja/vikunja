import type {TypeOf} from 'zod'
import {string} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'
import {FilterSchema} from './common/filter'

import {AbstractSchema} from './abstract'
import {UserSchema} from './user'

// FIXME: is it correct that this extends the Abstract Schema?
export const SavedFilterSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	title: string().default(''),
	description: string().default(''),
	filters: FilterSchema,

	owner: UserSchema,
	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
})

export type SavedFilter = TypeOf<typeof SavedFilterSchema> 
