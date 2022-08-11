import type {TypeOf} from 'zod'

import {IdSchema} from './common/id'

import {AbstractSchema} from './abstract'
import {ListSchema} from './list'

export const ListDuplicationSchema = AbstractSchema.extend({
	listId: IdSchema.default(0),
	namespaceId: IdSchema.default(0), // INamespace['id'],
	list: ListSchema,
})

export type ListDuplication = TypeOf<typeof ListDuplicationSchema> 