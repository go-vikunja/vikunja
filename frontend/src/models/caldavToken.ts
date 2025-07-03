import AbstractModel from './abstractModel'

import type {ICaldavToken} from '@/modelTypes/ICaldavToken'

export default class CaldavTokenModel extends AbstractModel<ICaldavToken> implements ICaldavToken {
	id: number = 0
	created: Date = new Date()

	constructor(data: Partial<ICaldavToken>) {
		super()
		this.assignData(data)
		
		if (this.created) {
			this.created = new Date(this.created)
		}
	}
}
