import AbstractModel from './abstractModel'

export default class PasswordUpdateModel extends AbstractModel {
	newPassword: string
	oldPassword: string

	defaults() {
		return {
			newPassword: '',
			oldPassword: '',
		}
	}
}