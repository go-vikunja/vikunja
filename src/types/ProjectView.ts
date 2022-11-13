export const PROJECT_VIEWS = {
	LIST: 'list',
	GANTT: 'gantt',
	TABLE: 'table',
	KANBAN: 'kanban',
} as const

export type ProjectView = typeof PROJECT_VIEWS[keyof typeof PROJECT_VIEWS]
