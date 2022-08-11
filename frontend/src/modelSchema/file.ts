import type {TypeOf} from 'zod'
import {object, number, string} from 'zod'

import {IdSchema} from './common/id'
import {DateSchema} from './common/date'

export const FileSchema = object({
  id: IdSchema.default(0),
  mime: string().default(''),
  name: string().default(''),
  size: number().default(0),
  created: DateSchema.nullable(),
})

export type File = TypeOf<typeof FileSchema> 
