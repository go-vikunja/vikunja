import type {TypeOf} from 'zod'
import {boolean, number, string, array, any} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'
import {SubscriptionSchema} from './subscription'
import {TaskSchema} from './task'
import {UserSchema} from './user'

export const ListSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	hash: string().default(''),
	description: string().default(''),
	owner: UserSchema,
	tasks: array(TaskSchema),
	namespaceId: IdSchema.default(0), // INamespace['id'],
	isArchived: boolean().default(false),
	hexColor: string().default(''),
	identifier: string().default(''),
	backgroundInformation: any().nullable().default(null), // FIXME: what is this for?
	isFavorite: boolean().default(false),
	subscription: SubscriptionSchema.nullable(),
	position: number().default(0),
	backgroundBlurHash: string().default(''),
	
	created: DateSchema.nullable(),
	updated: DateSchema.nullable(),
})

export type List = TypeOf<typeof ListSchema> 
