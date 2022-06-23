import AbstractModel from './abstractModel'
import UserModel from './user'
import type TaskModel from './task'

export const RELATION_KINDS = [
	'subtask',
	'parenttask',
	'related',
	'duplicates',
	'blocking',
	'blocked',
	'precedes',
	'follows',
	'copiedfrom',
	'copiedto',
  ] as const

export type RelationKind = typeof RELATION_KINDS[number]

export default class TaskRelationModel extends AbstractModel {
	id: number
	otherTaskId: TaskModel['id']
	taskId: TaskModel['id']
	relationKind: RelationKind

	createdBy: UserModel
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