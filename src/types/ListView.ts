export const LIST_VIEWS = {
	LIST: 'list',
	GANTT: 'gantt',
	TABLE: 'table',
	KANBAN: 'kanban',
} as const

export type ListView = typeof LIST_VIEWS[keyof typeof LIST_VIEWS]
