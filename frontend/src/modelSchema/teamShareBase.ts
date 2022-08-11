import type {TypeOf} from 'zod'
import {nativeEnum} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'

import {RIGHTS} from '@/constants/rights'

export const TeamShareBaseSchema = AbstractSchema.extend({
	teamId: IdSchema.default(0), // ITeam['id']
	right: nativeEnum(RIGHTS).default(RIGHTS.READ),

	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
})

export type ITeamShareBase = TypeOf<typeof TeamShareBaseSchema> 
