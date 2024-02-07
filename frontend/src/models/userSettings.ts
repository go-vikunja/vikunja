import AbstractModel from './abstractModel'

import type {IFrontendSettings, IUserSettings} from '@/modelTypes/IUserSettings'
import {getBrowserLanguage} from '@/i18n'
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
	language = getBrowserLanguage() 
	frontendSettings: IFrontendSettings = {
		playSoundWhenDone: true,
		quickAddMagicMode: PrefixMode.Default,
		colorSchema: 'auto',
	}

	constructor(data: Partial<IUserSettings> = {}) {
		super()
		this.assignData(data)
	}
}