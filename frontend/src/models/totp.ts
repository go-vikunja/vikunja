import AbstractModel from './abstractModel'

import type {ITotp} from '@/modelTypes/ITotp'

export default class TotpModel extends AbstractModel<ITotp> implements ITotp {
	secret = ''
	enabled = false
	url = ''

	constructor(data: Partial<ITotp> = {}) {
		super()
		this.assignData(data)
	}
}
