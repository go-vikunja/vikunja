import AbstractService from './abstractService'

export default class PasswordUpdateService extends AbstractService {
	constructor() {
		super({
			update: '/user/password',
		})
	}
}