export function findIndexById<T extends {id: string | number}>(array : T[], id : string | number) {
	return array.findIndex(({id: currentId}) => currentId === id)
}

export function findById<T extends {id: string | number}>(array : T[], id : string | number) {
	return array.find(({id: currentId}) => currentId === id)
}

interface ObjectWithId {
	id: number
}

export function includesById(array: ObjectWithId[], id: string | number) {
	return array.some(({id: currentId}) => currentId === id)
}

// https://github.com/you-dont-need/You-Dont-Need-Lodash-Underscore#_isnil
export function isNil(value: unknown) {
	return value == null
}

export function omitBy(obj: Record<string, unknown>, check: (value: unknown) => boolean) {
	if (isNil(obj)) {
		return {}
	}

	return Object.fromEntries(
		Object.entries(obj).filter(([, value]) => !check(value)),
	)
}

// Route parameter utilities for type safety
export function getRouteParamAsString(param: string | string[] | undefined): string | undefined {
	if (Array.isArray(param)) {
		return param[0]
	}
	return param
}

export function getRouteParamAsNumber(param: string | string[] | undefined): number | undefined {
	const stringParam = getRouteParamAsString(param)
	if (!stringParam) {
		return undefined
	}
	const num = parseInt(stringParam, 10)
	return isNaN(num) ? undefined : num
}
