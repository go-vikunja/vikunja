export const dateRanges = {
	// Format: 
	// Key is the title, as a translation string, the first entry of the value array 
	// is the "from" date, the second one is the "to" date.
	'today': ['now/d', 'now/d+1d'],

	'lastWeek': ['now/w-1w', 'now/w-2w'],
	'thisWeek': ['now/w', 'now/w+1w'],
	'restOfThisWeek': ['now', 'now/w+1w'],
	'nextWeek': ['now/w+1w', 'now/w+2w'],
	'next7Days': ['now', 'now+7d'],

	'lastMonth': ['now/M-1M', 'now/M-2M'],
	'thisMonth': ['now/M', 'now/M+1M'],
	'restOfThisMonth': ['now', 'now/M+1M'],
	'nextMonth': ['now/M+1M', 'now/M+2M'],
	'next30Days': ['now', 'now+30d'],
	
	'thisYear': ['now/y', 'now/y+1y'],
	'restOfThisYear': ['now', 'now/y+1y'],
}
