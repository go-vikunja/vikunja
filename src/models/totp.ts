import AbstractModel, { type IAbstract } from './abstractModel'

export interface ITotp extends IAbstract {
	secret: string
	enabled: boolean
	url: string
}

export default class TotpModel extends AbstractModel implements ITotp {
	secret = ''
	enabled = false
	url = ''

	constructor(data: Partial<ITotp>) {
		super()
		this.assignData(data)
	}
}