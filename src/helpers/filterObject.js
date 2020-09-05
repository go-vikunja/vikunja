export const filterObject = (obj, fn) => {
	let key

	for (key in obj) {
		if (fn(obj[key])) {
			return key
		}
	}
	return null
}
