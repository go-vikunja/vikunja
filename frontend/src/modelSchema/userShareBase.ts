import type {TypeOf} from 'zod'
import {nativeEnum} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'

import {RIGHTS} from '@/constants/rights'

export const UserShareBaseSchema = AbstractSchema.extend({
	userId: IdSchema, // FIXME: default of model is `''`
	right: nativeEnum(RIGHTS).default(RIGHTS.READ),

	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
})

export type TeamMember = TypeOf<typeof UserShareBaseSchema> 
