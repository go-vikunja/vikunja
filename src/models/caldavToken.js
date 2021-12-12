import AbstractModel from './abstractModel'

export default class CaldavTokenModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.created = new Date(this.created)
	}

	defaults() {
		return {
			id: 0,
			created: null,
		}
	}
}