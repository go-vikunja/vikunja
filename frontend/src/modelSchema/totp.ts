import type {TypeOf} from 'zod'
import {string, boolean} from 'zod'

import {AbstractSchema} from './abstract'

export const TotpSchema = AbstractSchema.extend({
  secret: string().default(''),
  enabled: boolean().default(false),
  url: string().url().default(''),
})

export type Totp = TypeOf<typeof TotpSchema> 
