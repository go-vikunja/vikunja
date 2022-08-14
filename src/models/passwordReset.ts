import AbstractModel, { type IAbstract } from './abstractModel'

export interface IPasswordReset extends IAbstract {
	token: string
	newPassword: string
	email: string
}

export default class PasswordResetModel extends AbstractModel implements IPasswordReset {
	token: string
	newPassword!: string
	email!: string

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