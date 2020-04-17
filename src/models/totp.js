import AbstractModel from './abstractModel'

export default class TotpModel extends AbstractModel {
	defaults() {
		return {
			secret: '',
			enabled: false,
			url: '',
		}
	}
}