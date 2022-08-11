import type {TypeOf} from 'zod'

import {object, string} from 'zod'

export const TokenSchema = object({
	token: string(),
})

export type IToken = TypeOf<typeof TokenSchema> 