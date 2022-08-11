import type {TypeOf} from 'zod'
import {object, nativeEnum} from 'zod'

import {RIGHTS} from '@/constants/rights'

export const AbstractSchema = object({
	maxRight: nativeEnum(RIGHTS).nullable(),
})

export type IAbstract = TypeOf<typeof AbstractSchema>