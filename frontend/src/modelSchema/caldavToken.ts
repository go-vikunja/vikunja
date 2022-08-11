
import type {TypeOf} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

import {AbstractSchema} from './abstract'

export const CaldavTokenSchema = AbstractSchema.extend({
  id: IdSchema,
  created: DateSchema,
})

export type CaldavToken = TypeOf<typeof CaldavTokenSchema> 
