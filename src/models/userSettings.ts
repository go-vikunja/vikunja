
import AbstractModel from './abstractModel'

import type {IUserSettings} from '@/modelTypes/IUserSettings'
import type {IList} from '@/modelTypes/IList'

export default class UserSettingsModel extends AbstractModel<IUserSettings> implements IUserSettings {
	name = ''
	emailRemindersEnabled = true
	discoverableByName = false
	discoverableByEmail = false
	overdueTasksRemindersEnabled = true
	defaultListId: undefined | IList['id'] = undefined
	weekStart: IUserSettings['weekStart'] = 0
	timezone = ''

	constructor(data: Partial<IUserSettings>) {
		super()
		this.assignData(data)
	}
}