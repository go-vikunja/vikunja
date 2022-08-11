import type {TypeOf} from 'zod'
import {string} from 'zod'

import {AbstractSchema} from './abstract'

export const EmailUpdateSchema = AbstractSchema.extend({
  newEmail: string().email().default(''),
  password: string().default(''),
})

export type EmailUpdate = TypeOf<typeof EmailUpdateSchema> 
