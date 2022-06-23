
import AbstractModel from './abstractModel'
import type ListModel from './list'

export default class UserSettingsModel extends AbstractModel {
	name: string
	emailRemindersEnabled: boolean
	discoverableByName: boolean
	discoverableByEmail: boolean
	overdueTasksRemindersEnabled: boolean
	defaultListId: undefined | ListModel['id']
	weekStart: 0 | 1 | 2 | 3 | 4 | 5 | 6
	timezone: string

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