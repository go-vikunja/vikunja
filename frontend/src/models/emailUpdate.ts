import AbstractModel from './abstractModel'

import type {IEmailUpdate} from '@/modelTypes/IEmailUpdate'

export default class EmailUpdateModel extends AbstractModel<IEmailUpdate> implements IEmailUpdate {
	newEmail = ''
	password = ''

	constructor(data : Partial<IEmailUpdate> = {}) {
		super()
		this.assignData(data)
	}
}
