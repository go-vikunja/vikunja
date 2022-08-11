import type {TypeOf} from 'zod'
import {nativeEnum, number, object, preprocess} from 'zod'

import {SECONDS_A_HOUR} from '@/constants/date'
import {REPEAT_TYPES, type IRepeatAfter} from '@/types/IRepeatAfter'

/**
 * Parses `repeatAfterSeconds` into a usable js object.
 */
 export function parseRepeatAfter(repeatAfterSeconds: number): IRepeatAfter {
	let repeatAfter: IRepeatAfter

	// if its dividable by SECONDS_A_DAY, its something with days, otherwise hours
	if (repeatAfterSeconds % SECONDS_A_DAY === 0) {
		repeatAfter = {type: REPEAT_TYPES.HOURS, amount: repeatAfterSeconds / SECONDS_A_HOUR}
	} else if (repeatAfterSeconds % SECONDS_A_WEEK === 0) {
		repeatAfter = {type: REPEAT_TYPES.WEEKS, amount: repeatAfterSeconds / SECONDS_A_WEEK}
	} else if (repeatAfterSeconds % SECONDS_A_MONTH === 0) {
		repeatAfter = {type: REPEAT_TYPES.MONTHS, amount: repeatAfterSeconds / SECONDS_A_MONTH}
	} else if (repeatAfterSeconds % SECONDS_A_YEAR === 0) {
		repeatAfter = {type: REPEAT_TYPES.YEARS, amount: repeatAfterSeconds / SECONDS_A_YEAR}
	} else {
		repeatAfter = {type: REPEAT_TYPES.DAYS, amount: repeatAfterSeconds / SECONDS_A_DAY}
	}
	return repeatAfter
}

export const RepeatsSchema = preprocess(
	(repeats: unknown) => {
		// Parses the "repeat after x seconds" from the task into a usable js object inside the task.
	
		if (typeof repeats !== 'number') {
			return repeats
		}
	
		return parseRepeatAfter(repeats)
	},
	object({
		type: nativeEnum(REPEAT_TYPES),
		amount: number().int(),
	}),
)

export type RepeatAfter = TypeOf<typeof RepeatsSchema>