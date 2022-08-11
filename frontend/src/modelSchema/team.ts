import type {TypeOf} from 'zod'
import {array, nativeEnum, string} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'
import {UserSchema} from './user'
import {TeamMemberSchema} from './teamMember'

import {RIGHTS} from '@/constants/rights'

export const TeamSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	name: string().default(''),
	description: string().default(''),
	members: array(TeamMemberSchema),
	right: nativeEnum(RIGHTS).default(RIGHTS.READ),

	createdBy: UserSchema, // FIXME: default was {},
	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
})

export type Team = TypeOf<typeof TeamSchema> 