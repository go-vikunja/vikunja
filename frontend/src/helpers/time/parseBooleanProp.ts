export function parseBooleanProp(booleanProp: string | undefined) {
	return (booleanProp === 'false' || booleanProp === '0')
		? false
		: Boolean(booleanProp)
}
