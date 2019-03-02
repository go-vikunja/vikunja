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
		this.sortTasks()
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

	////////
	// Helpers
	//////
	
	/**
	 * Sorts all tasks according to their due date
	 * @returns {this}
	 */
	sortTasks() {
		if (this.tasks === null || this.tasks === []) {
			return
		}
		return this.tasks.sort(function(a,b) {
			if (a.done < b.done)
				return -1
			if (a.done > b.done)
				return 1
			return 0
		})
	}
	
	/**
	 * Adds a task to the task array of this list. Usually only used when creating a new task
	 * @param task
	 */
	addTaskToList(task) {
		// If it's a subtask, add it to its parent, otherwise append it to the list of tasks
		if (task.parentTaskID === 0) {
			this.tasks.push(task)
		} else {
			for (const t in this.tasks) {
				if (this.tasks[t].id === task.parentTaskID) {
					this.tasks[t].subtasks.push(task)
					break
				}
			}
		}
		this.sortTasks()
	}
	
	/**
	 * Gets a task by its ID by looping through all tasks.
	 * @param id
	 * @returns {TaskModel}
	 */
	getTaskByID(id) {
		// TODO: Binary search?
		for (const t in this.tasks) {
			if (this.tasks[t].id === parseInt(id)) {
				return this.tasks[t]
			}
		}
		return {} // FIXME: This should probably throw something to make it clear to the user noting was found
	}

	/**
	 * Loops through all tasks and updates the one  with the id it has
	 * @param task
	 */
	updateTaskByID(task) {
		for (const t in this.tasks) {
			if (this.tasks[t].id === task.id) {
				this.tasks[t] = task
				break
			}

			if (this.tasks[t].id === task.parentTaskID) {
				for (const s in this.tasks[t].subtasks) {
					if (this.tasks[t].subtasks[s].id === task.id) {
						this.tasks[t].subtasks[s] = task
						break
					}
				}
			}
		}
		this.sortTasks()
	}
}