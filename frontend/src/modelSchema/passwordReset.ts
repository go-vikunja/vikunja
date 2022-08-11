import type {TypeOf} from 'zod'
import {string} from 'zod'

import {AbstractSchema} from './abstract'

// FIXME: is it correct that this extends the Abstract Schema?
export const PasswordResetSchema = AbstractSchema.extend({
  token: string().default(''),
  newPassword: string().default(''),
  email: string().email().default(''),
})

export type PasswordReset = TypeOf<typeof PasswordResetSchema> 
