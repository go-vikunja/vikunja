import AbstractModel from '@/models/abstractModel'

export interface IPasswordUpdate extends AbstractModel {
	newPassword: string
	oldPassword: string
}

export default class PasswordUpdateModel extends AbstractModel implements IPasswordUpdate {
	newPassword!: string
	oldPassword!: string

	defaults() {
		return {
			newPassword: '',
			oldPassword: '',
		}
	}
}