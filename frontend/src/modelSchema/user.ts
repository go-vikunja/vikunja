import type {TypeOf} from 'zod'
import {string} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'
import {UserSettingsSchema} from './userSettings'

export const UserSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	email: string().email().default(''),
	username: string().default(''),
	name: string().default(''),
	settings: UserSettingsSchema.nullable(),

	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
})

export type User = TypeOf<typeof UserSchema> 
