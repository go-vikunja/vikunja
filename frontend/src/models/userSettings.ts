import AbstractModel from './abstractModel'

import type {IFrontendSettings, IUserSettings} from '@/modelTypes/IUserSettings'
import {getBrowserLanguage} from '@/i18n'
import {PrefixMode} from '@/modules/parseTaskText'
import {DEFAULT_PROJECT_VIEW_SETTINGS} from '@/modelTypes/IProjectView'
import {PRIORITIES} from '@/constants/priorities'
import {DATE_DISPLAY} from '@/constants/dateDisplay'

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
		allowIconChanges: true,
		defaultView: DEFAULT_PROJECT_VIEW_SETTINGS.FIRST,
		minimumPriority: PRIORITIES.MEDIUM,
		dateDisplay: DATE_DISPLAY.RELATIVE,
	}
	extraSettingsLinks = {}

	constructor(data: Partial<IUserSettings> = {}) {
		super()
		this.assignData(data)
	}
}
