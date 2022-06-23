import AbstractModel from './abstractModel'

export default class TotpModel extends AbstractModel {
	secret: string
	enabled: boolean
	url: string

	defaults() {
		return {
			secret: '',
			enabled: false,
			url: '',
		}
	}
}