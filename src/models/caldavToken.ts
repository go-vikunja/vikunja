import AbstractModel from './abstractModel'

export interface ICaldavToken extends AbstractModel {
	id: number;
	created: Date;
}

export default class CaldavTokenModel extends AbstractModel implements ICaldavToken {
	id!: number
	created!: Date

	constructor(data? : Object) {
		super(data)
		
		this.id

		if (this.created) {
			this.created = new Date(this.created)
		}
	}
}