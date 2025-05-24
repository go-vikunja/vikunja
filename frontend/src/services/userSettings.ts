import type {IUserSettings} from '@/modelTypes/IUserSettings'
import AbstractService from './abstractService'

export default class UserSettingsService extends AbstractService<IUserSettings> {
	constructor() {
		super({
			update: '/user/settings/general',
		})
	}
}
