export function canNestProjectDeeper(level: number) {
	if (level < 2) {
		return true
	}

	return level >= 2 && window.PROJECT_INFINITE_NESTING_ENABLED
}