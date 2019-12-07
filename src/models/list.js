import AbstractModel from './abstractModel'
import TaskModel from './task'
import UserModel from './user'

export default class ListModel extends AbstractModel {
	
	constructor(data) {
		super(data)
		
		// Make all tasks to task models
		this.tasks = this.tasks.map(t => {
			return new TaskModel(t)
		})
		
		this.owner = new UserModel(this.owner)
	}
	
	// Default attributes that define the "empty" state.
	defaults() {
		return {
			id: 0,
			title: '',
			description: '',
			owner: UserModel,
			tasks: [],
			namespaceID: 0,
			
			created: 0,
			updated: 0,
		}
	}
}