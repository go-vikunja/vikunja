import AbstractModel from './abstractModel'

import type {ICaldavToken} from '@/modelTypes/ICaldavToken'

export default class CaldavTokenModel extends AbstractModel<ICaldavToken> implements ICaldavToken {
	id: number
	created: Date

	constructor(data: Partial<CaldavTokenModel>) {
		super()
		this.assignData(data)
		
		if (this.created) {
			this.created = new Date(this.created)
		}
	}
}
