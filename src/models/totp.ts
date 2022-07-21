import AbstractModel from './abstractModel'

export interface ITotp extends AbstractModel {
	secret: string
	enabled: boolean
	url: string
}

export default class TotpModel extends AbstractModel implements ITotp{
	declare secret: string
	declare enabled: boolean
	declare url: string

	defaults() {
		return {
			secret: '',
			enabled: false,
			url: '',
		}
	}
}