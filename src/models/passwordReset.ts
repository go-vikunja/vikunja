import AbstractModel, { type IAbstract } from './abstractModel'

export interface IPasswordReset extends IAbstract {
	token: string
	newPassword: string
	email: string
}

export default class PasswordResetModel extends AbstractModel implements IPasswordReset {
	token = ''
	newPassword = ''
	email = ''

	constructor(data: Partial<IPasswordReset>) {
		super()
		this.assignData(data)

		this.token = localStorage.getItem('passwordResetToken')
	}
}