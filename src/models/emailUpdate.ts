import AbstractModel, { type IAbstract } from './abstractModel'

interface IEmailUpdate extends IAbstract {
	newEmail: string
	password: string
}

export default class EmailUpdateModel extends AbstractModel implements IEmailUpdate {
	newEmail!: string
	password!: string

	defaults() {
		return {
			newEmail: '',
			password: '',
		}
	}
}