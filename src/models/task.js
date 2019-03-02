import AbstractModel from './abstractModel';
import UserModel from './user'

export default class TaskModel extends AbstractModel {
	
	constructor(data) {
		super(data)
		
		// Make date objects from timestamps
		this.dueDate = this.parseDateIfNessecary(this.dueDate)
		this.startDate = this.parseDateIfNessecary(this.startDate)
		this.endDate = this.parseDateIfNessecary(this.endDate)

		this.reminderDates = this.reminderDates.map(d => {
			return this.parseDateIfNessecary(d)
		})
		this.reminderDates.push(null) // To trigger the datepicker

		// Parse the repeat after into something usable
		this.parseRepeatAfter()
		
		// Parse the assignees into user models
		this.assignees = this.assignees.map(a => {
			return new UserModel(a)
		})
		this.createdBy = new UserModel(this.createdBy)
	}
	
	defaults() {
		return {
			id: 0,
			text: '',
			description: '',
			done: false,
			priority: 0,
			labels: [],
			assignees: [],
			
			dueDate: 0,
			startDate: 0,
			endDate: 0,
			repeatAfter: 0,
			reminderDates: [],
			subtasks: [],
			parentTaskID: 0,
			
			createdBy: UserModel,
			created: 0,
			updated: 0,
			
			listID: 0, // Meta, only used when creating a new task
			sortBy: 'duedate', // Meta, only used when listing all tasks
		}
	}
	
	/////////////////
	// Helper functions
	///////////////
	
	/**
	 * Makes a js date object from a unix timestamp (in seconds).
	 * @param unixTimestamp
	 * @returns {*}
	 */
	parseDateIfNessecary(unixTimestamp) {
		let dateobj = new Date(unixTimestamp * 1000)
		if (unixTimestamp === 0) {
			return null
		}
		return dateobj
	}

	/**
	 * Parses the "repeat after x seconds" from the task into a usable js object inside the task.
	 * This function should only be called from the constructor.
	 */
	parseRepeatAfter() {
		let repeatAfterHours = (this.repeatAfter / 60) / 60
		this.repeatAfter = {type: 'hours', amount: repeatAfterHours}

		// if its dividable by 24, its something with days, otherwise hours
		if (repeatAfterHours % 24 === 0) {
			let repeatAfterDays = repeatAfterHours / 24
			if (repeatAfterDays % 7 === 0) {
				this.repeatAfter.type = 'weeks'
				this.repeatAfter.amount = repeatAfterDays / 7
			} else if (repeatAfterDays % 30 === 0) {
				this.repeatAfter.type = 'months'
				this.repeatAfter.amount = repeatAfterDays / 30
			} else if (repeatAfterDays % 365 === 0) {
				this.repeatAfter.type = 'years'
				this.repeatAfter.amount = repeatAfterDays / 365
			} else {
				this.repeatAfter.type = 'days'
				this.repeatAfter.amount = repeatAfterDays
			}
		}
	}
}