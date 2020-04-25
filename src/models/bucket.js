import AbstractModel from './abstractModel'
import UserModel from './user'
import TaskModel from "./task";

export default class BucketModel extends AbstractModel {
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
			tasks: [],

			createdBy: null,
			created: null,
			updated: null,
		}
	}
}