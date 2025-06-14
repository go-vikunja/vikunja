import AbstractModel from './abstractModel'

import type {IPasswordReset} from '@/modelTypes/IPasswordReset'

export default class PasswordResetModel extends AbstractModel<IPasswordReset> implements IPasswordReset {
	token = ''
	newPassword = ''
	email = ''

	constructor(data: Partial<IPasswordReset> = {}) {
		super()
		this.assignData(data)

		if (data.token) {
			this.token = data.token
		}
	}
}
