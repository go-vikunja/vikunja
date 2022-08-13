import AbstractModel from './abstractModel'
import UserModel, { type IUser } from './user'
import type { ITask } from './task'

export const RELATION_KIND = {
	'SUBTASK': 'subtask',
	'PARENTTASK': 'parenttask',
	'RELATED': 'related',
	'DUPLICATES': 'duplicates',
	'BLOCKING': 'blocking',
	'BLOCKED': 'blocked',
	'PROCEDES': 'precedes',
	'FOLLOWS': 'follows',
	'COPIEDFROM': 'copiedfrom',
	'COPIEDTO': 'copiedto',
 } as const

export const RELATION_KINDS = [...Object.values(RELATION_KIND)] as const

export type RelationKind = typeof RELATION_KINDS[number]

export interface ITaskRelation extends AbstractModel {
	id: number
	otherTaskId: ITask['id']
	taskId: ITask['id']
	relationKind: RelationKind

	createdBy: IUser
	created: Date
}

export default class TaskRelationModel extends AbstractModel implements ITaskRelation {
	id!: number
	otherTaskId!: ITask['id']
	taskId!: ITask['id']
	relationKind!: RelationKind

	createdBy: IUser
	created: Date

	constructor(data) {
		super(data)
		this.createdBy = new UserModel(this.createdBy)
		this.created = new Date(this.created)
	}

	defaults() {
		return {
			id: 0,
			otherTaskId: 0,
			taskId: 0,
			relationKind: '',

			createdBy: UserModel,
			created: null,
		}
	}
}