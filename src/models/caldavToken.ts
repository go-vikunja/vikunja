import AbstractModel from './abstractModel'

export default class CaldavTokenModel extends AbstractModel {
	id = 0
	created : undefined | Date = undefined

	constructor(data? : Object) {
		super(data)

		if (this.created) {
			this.created = new Date(this.created)
		}
	}
}