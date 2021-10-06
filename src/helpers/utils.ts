export function findIndexById(array : [], id : string | number) {
	return array.findIndex(({id: currentId}) => currentId === id)
}

export function includesById(array: [], id: string | number) {
	return array.some(({id: currentId}) => currentId === id)
}

// https://github.com/you-dont-need/You-Dont-Need-Lodash-Underscore#_isnil
export function isNil(value: any) {
	return value == null
}

export function omitBy(obj: {}, check: (value: any) => Boolean): {} {
	if (isNil(obj)) {
		return {}
	}

	return Object.fromEntries(
		Object.entries(obj).filter(([, value]) => !check(value)),
	)
}