import type {TypeOf} from 'zod'
import {boolean, string, array} from 'zod'

import {IdSchema} from './common/id'
import {HexColorSchema} from './common/hexColor'

import {AbstractSchema} from './abstract'
import {ListSchema} from './list'
import {UserSchema} from './user'
import {SubscriptionSchema} from './subscription'

export const NamespaceSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	title: string().default(''),
	description: string().default(''),
	owner: UserSchema,
	lists: array(ListSchema),
	isArchived: boolean().default(false),
	hexColor: HexColorSchema.default(''),
	subscription: SubscriptionSchema.nullable(),
	
	created: IdSchema.nullable(),
	updated: IdSchema.nullable(),
})

export type Namespace = TypeOf<typeof NamespaceSchema> 
