import AbstractModel from './abstractModel'

export default class UserShareBaseModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			userID: 0,
			right: 0,
			
			created: null,
			updated: null,
		}
	}
}