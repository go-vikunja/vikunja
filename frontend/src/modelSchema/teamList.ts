import type {TypeOf} from 'zod'

import {IdSchema} from './common/id'

import {TeamShareBaseSchema} from './teamShareBase'

export const TeamListSchema = TeamShareBaseSchema.extend({
	listId: IdSchema.default(0), // IList['id']
})

export type TeamList = TypeOf<typeof TeamListSchema> 
