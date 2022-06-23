import AbstractModel from './abstractModel'
import UserModel from './user'
import TaskModel from './task'

export default class BucketModel extends AbstractModel {
	id: number
	title: string
	listId: number
	limit: number
	tasks: TaskModel[]
	isDoneBucket: boolean
	position: number
	
	createdBy: UserModel
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