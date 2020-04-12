import AbstractModel from './abstractModel'

/**
 * This class is a base class for common team sharing model.
 * It is extended in a way so it can be used for namespaces as well for lists.
 */
export default class TeamShareBaseModel extends AbstractModel {
	constructor(data) {
		super(data)
		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			teamId: 0,
			right: 0,
			
			created: null,
			updated: null
		}
	}
}