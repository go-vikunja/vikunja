import AbstractModel from './abstractModel'

export interface IPasswordReset {
	token: string
	newPassword: string
	email: string
}

export default class PasswordResetModel extends AbstractModel implements IPasswordReset {
	token: string
	declare newPassword: string
	declare email: string

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