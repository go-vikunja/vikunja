import type {TypeOf} from 'zod'
import {boolean, string, undefined, nativeEnum} from 'zod'

import {IdSchema} from './common/id'

import {AbstractSchema} from './abstract'

const WEEKDAYS = {
	MONDAY: 0,
	TUESDAY: 1,
	WEDNESDAY: 2,
	THURSDAY: 3,
	FRIDAY: 4,
	SATURDAY: 5,
	SUNDAY: 6,
} as const

export const UserSettingsSchema = AbstractSchema.extend({
	name: string().default(''),
	emailRemindersEnabled: boolean().default(true),
	discoverableByName: boolean().default(false),
	discoverableByEmail: boolean().default(false),
	overdueTasksRemindersEnabled: boolean().default(true),
	defaultListId: IdSchema.or(undefined()), // iListSchema['id'] // FIXME: shouldn't this be `null`?
	weekStart: nativeEnum(WEEKDAYS).default(WEEKDAYS.MONDAY),
	timezone: string().default(''),
})

export type IUserSettings = TypeOf<typeof UserSettingsSchema> 
