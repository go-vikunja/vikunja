export function parseBooleanProp(booleanProp: string) {
	return (booleanProp === 'false' || booleanProp === '0')
		? false
		:	Boolean(booleanProp)
}