import AbstractModel from './abstractModel'

export default class PasswordResetModel extends AbstractModel {
	token: string
	newPassword: string
	email: string

	constructor(data) {
		super(data)

		this.token = localStorage.getItem('passwordResetToken')
	}

	defaults() {
		return {
			token: '',
			newPassword: '',
			email: '',
		}
	}
}