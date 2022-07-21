import AbstractModel from './abstractModel'
import UserModel, { type IUser } from './user'
import TaskModel, { type ITask } from './task'

export interface IBucket extends AbstractModel {
	id: number
	title: string
	listId: number
	limit: number
	tasks: ITask[]
	isDoneBucket: boolean
	position: number
	
	createdBy: IUser
	created: Date
	updated: Date
}

export default class BucketModel extends AbstractModel implements IBucket {
	declare id: number
	declare title: string
	declare listId: number
	declare limit: number
	declare tasks: ITask[]
	declare isDoneBucket: boolean
	declare position: number
	
	createdBy: IUser
	created: Date
	updated: Date

	constructor(bucket) {
		super(bucket)

		this.tasks = this.tasks.map(t => new TaskModel(t))

		this.createdBy = new UserModel(this.createdBy)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			id: 0,
			title: '',
			listId: 0,
			limit: 0,
			tasks: [],
			isDoneBucket: false,
			position: 0,

			createdBy: null,
			created: null,
			updated: null,
		}
	}
}