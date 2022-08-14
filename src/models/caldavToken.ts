import AbstractModel, { type IAbstract } from './abstractModel'

export interface ICaldavToken extends IAbstract {
	id: number;
	created: Date;
}

export default class CaldavTokenModel extends AbstractModel implements ICaldavToken {
	id: number
	created: Date

	constructor(data? : Partial<CaldavTokenModel>) {
		super()
		this.assignData(data)
		
		if (this.created) {
			this.created = new Date(this.created)
		}
	}
}