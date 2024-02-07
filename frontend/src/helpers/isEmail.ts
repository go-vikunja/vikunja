export function isEmail(email: string): boolean {
	const format = /^.+@.+$/
	const match = email.match(format)

	return match === null ? false : match.length > 0
}
