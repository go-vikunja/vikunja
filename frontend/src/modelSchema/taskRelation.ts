import type {TypeOf} from 'zod'
import {nativeEnum} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'
import {UserSchema} from './user'

import {RELATION_KIND} from '@/types/IRelationKind'

export const TaskRelationSchema = AbstractSchema.extend({
	id: IdSchema.default(0),
	otherTaskId: IdSchema.default(0),
	taskId: IdSchema.default(0),
	relationKind: nativeEnum(RELATION_KIND).nullable().default(null), // FIXME: default value was empty string?
	
	createdBy: UserSchema,
	// FIXME: shouldn't the empty value of dates be `new Date()`
	// Because e.g. : `new Date(null)` => Thu Jan 01 1970 01:00:00 GMT+0100 (Central European Standard Time)
	created: DateSchema.nullable().default(null),
}) 

export type ITaskRelation = TypeOf<typeof TaskRelationSchema> 
