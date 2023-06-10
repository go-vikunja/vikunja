export const REMINDER_PERIOD_RELATIVE_TO_TYPES = {
	DUEDATE: 'due_date',
	STARTDATE: 'start_date',
	ENDDATE: 'end_date',
} as const

export type IReminderPeriodRelativeTo = typeof REMINDER_PERIOD_RELATIVE_TO_TYPES[keyof typeof REMINDER_PERIOD_RELATIVE_TO_TYPES]

