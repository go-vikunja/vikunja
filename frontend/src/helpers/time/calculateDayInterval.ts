type Day<T extends number = number> = T

export function calculateDayInterval(dateString: string, currentDay = (new Date().getDay())): Day {
	switch (dateString) {
		case 'today':
			return 0
		case 'tomorrow':
			return 1
		case 'nextMonday':
			// Monday is 1, so we calculate the distance to the next 1
			return (currentDay + (8 - currentDay * 2)) % 7
		case 'thisWeekend':
			// Saturday is 6 so we calculate the distance to the next 6
			return (6 - currentDay) % 6
		case 'laterThisWeek':
			if (currentDay === 5 || currentDay === 6 || currentDay === 0) {
				return 0
			}

			return 2
		case 'laterNextWeek':
			return calculateDayInterval('laterThisWeek', currentDay) + 7
		case 'nextWeek':
			return 7
		default:
			return 0
	}
}
