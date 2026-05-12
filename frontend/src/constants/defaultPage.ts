export const DEFAULT_PAGE = {
	OVERVIEW: 'overview',
	UPCOMING: 'upcoming',
	DEFAULT_PROJECT: 'defaultProject',
	LAST_VISITED: 'lastVisited',
} as const

export type DefaultPage = typeof DEFAULT_PAGE[keyof typeof DEFAULT_PAGE]
