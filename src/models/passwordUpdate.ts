import AbstractModel from './abstractModel'

export default class PasswordUpdateModel extends AbstractModel {
	defaults() {
		return {
			newPassword: '',
			oldPassword: '',
		}
	}
}