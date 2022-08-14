import AbstractModel, { type IAbstract } from '@/models/abstractModel'

export interface IPasswordUpdate extends IAbstract {
	newPassword: string
	oldPassword: string
}

export default class PasswordUpdateModel extends AbstractModel implements IPasswordUpdate {
	newPassword = ''
	oldPassword = ''

	constructor(data: Partial<IPasswordUpdate>) {
		super()
		this.assignData(data)
	}
}