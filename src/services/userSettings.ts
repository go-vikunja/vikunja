
import type { IUserSettings } from '@/models/userSettings'
import AbstractService from './abstractService'

export default class UserSettingsService extends AbstractService<IUserSettings> {
	constructor() {
		super({
			update: '/user/settings/general',
		})
	}
}