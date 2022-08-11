import type {TypeOf} from 'zod'
import {string} from 'zod'

import {DateSchema} from './common/date'
import {IdSchema} from './common/id'

import {AbstractSchema} from './abstract'
import {UserSchema} from './user'

export const SubscriptionSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	entity: string().default(''), // FIXME: correct type?
	entityId: IdSchema.default(0), // FIXME: correct type?
	user: UserSchema,

	created: DateSchema.nullable(),
})

export type Subscription = TypeOf<typeof SubscriptionSchema> 
