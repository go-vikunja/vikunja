import AbstractModel from './abstractModel'

import type {IUserSettings} from '@/modelTypes/IUserSettings'
import {getCurrentLanguage} from '@/i18n'

export default class UserSettingsModel extends AbstractModel<IUserSettings> implements IUserSettings {
	name = ''
	emailRemindersEnabled = true
	discoverableByName = false
	discoverableByEmail = false
	overdueTasksRemindersEnabled = true
	overdueTasksRemindersTime = undefined
	defaultListId = undefined
	weekStart = 0 as IUserSettings['weekStart']
	timezone = ''
	language = getCurrentLanguage()

	constructor(data: Partial<IUserSettings> = {}) {
		super()
		this.assignData(data)
	}
}