export const DATE_RANGES = {
	// Format: 
	// Key is the title, as a translation string, the first entry of the value array 
	// is the "from" date, the second one is the "to" date.
	'today': ['now/d', 'now/d+1d'],

	'lastWeek': ['now/w-1w', 'now/w'],
	'thisWeek': ['now/w', 'now/w+1w'],
	'restOfThisWeek': ['now', 'now/w+1w'],
	'nextWeek': ['now/w+1w', 'now/w+2w'],
	'next7Days': ['now', 'now+7d'],

	'lastMonth': ['now/M-1M', 'now/M'],
	'thisMonth': ['now/M', 'now/M+1M'],
	'restOfThisMonth': ['now', 'now/M+1M'],
	'nextMonth': ['now/M+1M', 'now/M+2M'],
	'next30Days': ['now', 'now+30d'],
	
	'thisYear': ['now/y', 'now/y+1y'],
	'restOfThisYear': ['now', 'now/y+1y'],
} as const

export const DATE_VALUES = {
	'now': 'now',
	'startOfToday': 'now/d',
	'endOfToday': 'now/d+1d',

	'beginningOflastWeek': 'now/w-1w',
	'endOfLastWeek': 'now/w',
	'beginningOfThisWeek': 'now/w',
	'endOfThisWeek': 'now/w+1w',
	'startOfNextWeek': 'now/w+1w',
	'endOfNextWeek': 'now/w+2w',
	'in7Days': 'now+7d',

	'beginningOfLastMonth': 'now/M-1M',
	'endOfLastMonth': 'now/M',
	'startOfThisMonth': 'now/M',
	'endOfThisMonth': 'now/M+1M',
	'startOfNextMonth': 'now/M+1M',
	'endOfNextMonth': 'now/M+2M',
	'in30Days': 'now+30d',

	'startOfThisYear': 'now/y',
	'endOfThisYear': 'now/y+1y',
} as const
