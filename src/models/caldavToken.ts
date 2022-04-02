import AbstractModel from './abstractModel'

export default class CaldavTokenModel extends AbstractModel {
	constructor(data? : Object) {
		super(data)
		
		/** @type {number} */
		this.id

		if (this.created) {
			/** @type {Date} */
			this.created = new Date(this.created)
		}
	}
}