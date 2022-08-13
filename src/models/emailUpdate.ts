import AbstractModel from './abstractModel'

interface IEmailUpdate extends AbstractModel {
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