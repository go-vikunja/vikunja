import AbstractModel, { type IAbstract } from '@/models/abstractModel'

export interface IPasswordUpdate extends IAbstract {
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