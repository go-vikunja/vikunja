import AbstractService from './abstractService'

export default class EmailUpdateService extends AbstractService {
	constructor() {
		super({
			update: '/user/settings/email',
		})
	}
}
