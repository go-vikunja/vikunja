
import AbstractModel from './abstractModel'
import type { IList } from './list'

export interface IUserSettings extends AbstractModel {
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
	name!: string
	emailRemindersEnabled!: boolean
	discoverableByName!: boolean
	discoverableByEmail!: boolean
	overdueTasksRemindersEnabled!: boolean
	defaultListId!: undefined | IList['id']
	weekStart!: 0 | 1 | 2 | 3 | 4 | 5 | 6
	timezone!: string

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