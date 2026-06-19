import type {IAbstract} from './IAbstract'

export interface ITimeEntry extends IAbstract {
	id: number
	userId: number
	// Exactly one of taskId / projectId is set (0 means unset).
	taskId: number
	projectId: number
	startTime: Date
	// null while the live timer is running.
	endTime: Date | null
	comment: string

	created: Date
	updated: Date
}
