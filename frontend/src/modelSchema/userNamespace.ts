import type {TypeOf} from 'zod'

import {IdSchema} from './common/id'

import {UserShareBaseSchema} from './userShareBase'

export const UserNamespaceSchema = UserShareBaseSchema.extend({
	namespaceId: IdSchema.default(0), // INamespace['id']
})

export type IUserNamespace = TypeOf<typeof UserNamespaceSchema> 
