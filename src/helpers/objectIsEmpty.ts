// https://stackoverflow.com/a/32108184/10924593
export function objectIsEmpty(obj: Record<string, unknown>): boolean {
	return obj
		&& Object.keys(obj).length === 0
		&& Object.getPrototypeOf(obj) === Object.prototype
}