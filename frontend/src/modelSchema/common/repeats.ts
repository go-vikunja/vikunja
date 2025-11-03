import {SECONDS_A_HOUR} from '@/constants/date'
import { REPEAT_TYPES, type IRepeatAfter } from '@/types/IRepeatAfter'
import { nativeEnum, number, object, preprocess } from 'zod'

/**
 * Parses `repeatAfterSeconds` into a usable js object.
 */
export function parseRepeatAfter(repeatAfterSeconds: number): IRepeatAfter {
	let repeatAfter: IRepeatAfter = {type: 'hours', amount: repeatAfterSeconds / SECONDS_A_HOUR}

	// if its dividable by 24, its something with days, otherwise hours
	if (repeatAfterSeconds % SECONDS_A_DAY === 0) {
		if (repeatAfterSeconds % SECONDS_A_WEEK === 0) {
			repeatAfter = {type: 'weeks', amount: repeatAfterSeconds / SECONDS_A_WEEK}
		} else if (repeatAfterSeconds % SECONDS_A_MONTH === 0) {
			repeatAfter = {type:'months', amount: repeatAfterSeconds / SECONDS_A_MONTH}
		} else if (repeatAfterSeconds % SECONDS_A_YEAR === 0) {
			repeatAfter = {type: 'years', amount: repeatAfterSeconds / SECONDS_A_YEAR}
		} else {
			repeatAfter = {type: 'days', amount: repeatAfterSeconds / SECONDS_A_DAY}
		}
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
