import AbstractModel from './abstractModel';
import UserModel from './user'
import LabelModel from "./label";

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
		
		this.labels = this.labels.map(l => {
			return new LabelModel(l)
		})

		// Set the default color
		if (this.hexColor === '') {
			this.hexColor = '198CFF'
		}
		if (this.hexColor.substring(0, 1) !== '#') {
			this.hexColor = '#' + this.hexColor
		}
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
			hexColor: '',

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

	/**
	 * Checks if the hexColor of a task is dark.
	 * @returns {boolean}
	 */
	hasDarkColor() {
		if (this.hexColor === '#') {
			return true // Defaults to dark
		}

		let rgb = parseInt(this.hexColor.substring(1, 7), 16);   // convert rrggbb to decimal
		let r = (rgb >> 16) & 0xff;  // extract red
		let g = (rgb >>  8) & 0xff;  // extract green
		let b = (rgb >>  0) & 0xff;  // extract blue

		// luma will be a value 0..255 where 0 indicates the darkest, and 255 the brightest
		let luma = 0.2126 * r + 0.7152 * g + 0.0722 * b; // per ITU-R BT.709
		return luma > 128
	}
}