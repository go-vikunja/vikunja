import AbstractModel from './abstractModel'

import type {IFrontendSettings, IUserSettings} from '@/modelTypes/IUserSettings'
import {getCurrentLanguage} from '@/i18n'
import {PrefixMode} from '@/modules/parseTaskText'

export default class UserSettingsModel extends AbstractModel<IUserSettings> implements IUserSettings {
	name = ''
	emailRemindersEnabled = true
	discoverableByName = false
	discoverableByEmail = false
	overdueTasksRemindersEnabled = true
	overdueTasksRemindersTime = undefined
	defaultProjectId = undefined
	weekStart = 0 as IUserSettings['weekStart']
	timezone = ''
	language = getCurrentLanguage()
	frontendSettings: IFrontendSettings = {
		playSoundWhenDone: true,
		quickAddMagicMode: PrefixMode.Default,
	}

	constructor(data: Partial<IUserSettings> = {}) {
		super()
		this.assignData(data)
	}
}