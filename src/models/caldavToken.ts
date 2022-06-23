import AbstractModel from './abstractModel'

export default class CaldavTokenModel extends AbstractModel {
	id: number
	created: Date

	constructor(data? : Object) {
		super(data)
		
		this.id

		if (this.created) {
			this.created = new Date(this.created)
		}
	}
}