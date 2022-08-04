export const TASK_REPEAT_MODES = {
	'REPEAT_MODE_DEFAULT': 0,
	'REPEAT_MODE_MONTH': 1,
	'REPEAT_MODE_FROM_CURRENT_DATE': 2,
} as const

export type IRepeatMode = typeof TASK_REPEAT_MODES[keyof typeof TASK_REPEAT_MODES] 
