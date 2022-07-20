
import AbstractModel from './abstractModel'
import type { IList } from './list'

export interface IUserSettings {
	name: string
	emailRemindersEnabled: boolean
	discoverableByName: boolean
	discoverableByEmail: boolean
	overdueTasksRemindersEnabled: boolean
	defaultListId: undefined | IList['id']
	weekStart: 0 | 1 | 2 | 3 | 4 | 5 | 6
	timezone: string
}

export default class UserSettingsModel extends AbstractModel implements IUserSettings {
	declare name: string
	declare emailRemindersEnabled: boolean
	declare discoverableByName: boolean
	declare discoverableByEmail: boolean
	declare overdueTasksRemindersEnabled: boolean
	declare defaultListId: undefined | IList['id']
	declare weekStart: 0 | 1 | 2 | 3 | 4 | 5 | 6
	declare timezone: string

	defaults() {
		return {
			name: '',
			emailRemindersEnabled: true,
			discoverableByName: false,
			discoverableByEmail: false,
			overdueTasksRemindersEnabled: true,
			defaultListId: undefined,
			weekStart: 0,
			timezone: '',
		}
	}
}