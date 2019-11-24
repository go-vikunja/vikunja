import AbstractService from './abstractService'
import TaskModel from '../models/task'

export default class TaskService extends AbstractService {
	constructor() {
		super({
			create: '/lists/{listID}',
			getAll: '/tasks/all',
			get: '/tasks/{id}',
			update: '/tasks/{id}',
			delete: '/tasks/{id}',
		});
	}
	
	modelFactory(data) {
		return new TaskModel(data)
	}

	beforeUpdate(model) {
		return this.processModel(model)
	}

	beforeCreate(model) {
		return this.processModel(model)
	}

	processModel(model) {
		// Ensure the listID is an int
		model.listID = Number(model.listID)

		// Convert the date in a unix timestamp
		model.dueDate = model.dueDate !== null ? Math.round(+new Date(model.dueDate) / 1000) : model.dueDate
		model.startDate = model.startDate !== null ? Math.round(+new Date(model.startDate) / 1000): model.startDate
		model.endDate = model.endDate !== null ? Math.round(+new Date(model.endDate) / 1000) : model.endDate

		// remove all nulls, these would create empty reminders
		for (const index in model.reminderDates) {
			if (model.reminderDates[index] === null) {
				model.reminderDates.splice(index, 1)
			}
		}

		// Make normal timestamps from js dates
		if(model.reminderDates.length > 0) {
			model.reminderDates = model.reminderDates.map(r => {
				return Math.round(+new Date(r) / 1000)
			})
		}

		// Make the repeating amount to seconds
		let repeatAfterSeconds = 0
		if (model.repeatAfter.amount !== null || model.repeatAfter.amount !== 0) {
			switch (model.repeatAfter.type) {
				case 'hours':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60
					break
				case 'days':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60 * 24
					break
				case 'weeks':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60 * 24 * 7
					break
				case 'months':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60 * 24 * 30
					break
				case 'years':
					repeatAfterSeconds = model.repeatAfter.amount * 60 * 60 * 24 * 365
					break
			}
		}
		model.repeatAfter = repeatAfterSeconds

		if (model.hexColor.substring(0, 1) === '#') {
			model.hexColor = model.hexColor.substring(1, 7)
		}

		// Do the same for all related tasks
		Object.keys(model.related_tasks).forEach(relationKind  => {
			model.related_tasks[relationKind] = model.related_tasks[relationKind].map(t => {
				return this.processModel(t)
			})
		})

		return model
	}
}