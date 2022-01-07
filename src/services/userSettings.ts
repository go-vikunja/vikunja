
import AbstractService from './abstractService'

export default class UserSettingsService extends AbstractService {
	constructor() {
		super({
			update: '/user/settings/general',
		})
	}
}