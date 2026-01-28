/**
 * Parses scopes from a query parameter string in the format "group:permission,group:permission"
 * @param scopesParam - The raw scopes query parameter value
 * @returns An object mapping group names to arrays of permissions
 */
export function parseScopesFromQuery(scopesParam: string | null | undefined): Record<string, string[]> {
	if (!scopesParam) return {}

	const result: Record<string, string[]> = {}
	const pairs = scopesParam.split(',').map(s => s.trim()).filter(Boolean)

	for (const pair of pairs) {
		const [group, permission] = pair.split(':').map(s => s.trim())
		if (group && permission) {
			if (!result[group]) {
				result[group] = []
			}
			result[group].push(permission)
		}
	}

	return result
}
