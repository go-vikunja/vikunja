import AbstractModel from './abstractModel'

export default class EmailUpdateModel extends AbstractModel {
	defaults() {
		return {
			newEmail: '',
			passwort: '',
		}
	}
}