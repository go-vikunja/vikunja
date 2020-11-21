
import AbstractService from './abstractService'

export default class UserNameService extends AbstractService {
	constructor() {
		super({
			update: '/user/settings/name',
		})
	}
}