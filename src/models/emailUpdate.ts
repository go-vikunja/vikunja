import AbstractModel, { type IAbstract } from './abstractModel'

interface IEmailUpdate extends IAbstract {
	newEmail: string
	password: string
}

export default class EmailUpdateModel extends AbstractModel implements IEmailUpdate {
	newEmail = ''
	password = ''

	constructor(data : Partial<IEmailUpdate>) {
		super()
		this.assignData(data)
	}
}