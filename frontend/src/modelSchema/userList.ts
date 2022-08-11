import type {TypeOf} from 'zod'

import {IdSchema} from './common/id'

import {UserShareBaseSchema} from './userShareBase'

export const UserListSchema = UserShareBaseSchema.extend({
	listId: IdSchema.default(0), // IList['id']
})

export type IUserList = TypeOf<typeof UserListSchema> 
