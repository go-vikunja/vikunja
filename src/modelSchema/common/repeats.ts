import { REPEAT_TYPES, type IRepeatAfter } from '@/types/IRepeatAfter'
import { nativeEnum, number, object, preprocess } from 'zod'

export const RepeatsSchema = preprocess(
	(repeats: unknown) => {
		// Parses the "repeat after x seconds" from the task into a usable js object inside the task.
	
		if (typeof repeats !== 'number') {
			return repeats
		}
	
		const repeatAfterHours = (repeats / 60) / 60
		
		const repeatAfter : IRepeatAfter = {
			type: 'hours',
			amount: repeatAfterHours,
		}
	
		// if its dividable by 24, its something with days, otherwise hours
		if (repeatAfterHours % 24 === 0) {
			const repeatAfterDays = repeatAfterHours / 24
			if (repeatAfterDays % 7 === 0) {
				repeatAfter.type = 'weeks'
				repeatAfter.amount = repeatAfterDays / 7
			} else if (repeatAfterDays % 30 === 0) {
				repeatAfter.type = 'months'
				repeatAfter.amount = repeatAfterDays / 30
			} else if (repeatAfterDays % 365 === 0) {
				repeatAfter.type = 'years'
				repeatAfter.amount = repeatAfterDays / 365
			} else {
				repeatAfter.type = 'days'
				repeatAfter.amount = repeatAfterDays
			}
		}
	
		return repeatAfter
	},
	object({
		type: nativeEnum(REPEAT_TYPES),
		amount: number().int(),
	}),
)