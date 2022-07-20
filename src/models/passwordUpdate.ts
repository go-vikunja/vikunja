import AbstractModel from '@/models/abstractModel'

export interface IPasswordUpdate {
	newPassword: string
	oldPassword: string
}

export default class PasswordUpdateModel extends AbstractModel implements IPasswordUpdate {
	declare newPassword: string
	declare oldPassword: string

	defaults() {
		return {
			newPassword: '',
			oldPassword: '',
		}
	}
}