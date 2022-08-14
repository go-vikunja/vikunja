import AbstractModel, { type IAbstract } from './abstractModel'

export interface ITotp extends IAbstract {
	secret: string
	enabled: boolean
	url: string
}

export default class TotpModel extends AbstractModel implements ITotp{
	secret!: string
	enabled!: boolean
	url!: string

	defaults() {
		return {
			secret: '',
			enabled: false,
			url: '',
		}
	}
}