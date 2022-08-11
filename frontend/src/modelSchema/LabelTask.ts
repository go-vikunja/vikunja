import type {TypeOf} from 'zod'

import {IdSchema} from './common/id'

import {AbstractSchema} from './abstract'

export const LabelTaskSchema = AbstractSchema.extend({
  id: IdSchema.nullable(),
  taskId: IdSchema.nullable(),
  labelId: IdSchema.nullable(),
})

export type LabelTask = TypeOf<typeof LabelTaskSchema>