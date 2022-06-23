import AbstractModel from './abstractModel'

export default class EmailUpdateModel extends AbstractModel {
	newEmail: string
	password: string

	defaults() {
		return {
			newEmail: '',
			password: '',
		}
	}
}