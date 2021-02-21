export const parseDateOrNull = date => {
	if (date && !date.startsWith('0001')) {
		return new Date(date)
	}
	return null
}
