import type {TypeOf} from 'zod'

import {IdSchema} from './common/id'

import {TeamShareBaseSchema} from './teamShareBase'

export const TeamNamespaceSchema = TeamShareBaseSchema.extend({
	namespaceId: IdSchema.default(0), // INamespace['id']
})

export type ITeamNamespace = TypeOf<typeof TeamNamespaceSchema> 
