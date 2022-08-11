import type {TypeOf} from 'zod'
import {number, string, nativeEnum} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'
import {UserSchema} from './user'

import {RIGHTS} from '@/constants/rights'

export const LinkShareSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	hash: string().default(''),
	right: nativeEnum(RIGHTS).default(RIGHTS.READ),
	sharedBy: UserSchema,
	sharingType: number().default(0), // FIXME: use correct numbers
	listId: IdSchema.default(0),
	name: string().default(''),
	password: string().default(''),
	
	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
})

export type LinkShare = TypeOf<typeof LinkShareSchema> 
