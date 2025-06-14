import type {IAbstract} from './IAbstract'
import type {IUser} from './IUser'
import type {ITask} from './ITask'

import type {IRelationKind} from '@/types/IRelationKind'

export interface ITaskRelation extends IAbstract {
	id: number
	otherTaskId: ITask['id']
	taskId: ITask['id']
	relationKind: IRelationKind

	createdBy: IUser
	created: Date
}
