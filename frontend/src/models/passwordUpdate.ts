import AbstractModel from './abstractModel'

import type {IPasswordUpdate} from '@/modelTypes/IPasswordUpdate'

export default class PasswordUpdateModel extends AbstractModel<IPasswordUpdate> implements IPasswordUpdate {
	newPassword = ''
	oldPassword = ''

	constructor(data: Partial<IPasswordUpdate> = {}) {
		super()
		this.assignData(data)
	}
}
