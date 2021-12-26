export function hourToSalutation(now: Date = new Date()): String {
	const hours = new Date(now).getHours()

	if (hours < 5) {
		return 'Night'
	}

	if (hours < 11) {
		return 'Morning'
	}

	if (hours < 18) {
		return 'Day'
	}

	if (hours < 23) {
		return 'Evening'
	}

	return 'Night'
}
