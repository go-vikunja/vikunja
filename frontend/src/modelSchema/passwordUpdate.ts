import type {TypeOf} from 'zod'
import {string} from 'zod'

import {AbstractSchema} from './abstract'

// FIXME: is it correct that this extends the Abstract Schema?
export const PasswordUpdateSchema = AbstractSchema.extend({
  newPassword: string().default(''),
  oldPassword: string().default(''),
}).refine((data) => data.newPassword === data.oldPassword, {
	message: 'Passwords don\'t match',
	path: ['confirm'], // path of error
})

export type PasswordUpdate = TypeOf<typeof PasswordUpdateSchema> 
