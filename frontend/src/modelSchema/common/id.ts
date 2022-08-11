import {number, preprocess} from 'zod'

export const IdSchema = preprocess(
	(value: unknown) => Number(value),
	number().positive().int(),
)