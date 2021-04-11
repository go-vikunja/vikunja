
import AbstractModel from './abstractModel'

export default class UserSettingsModel extends AbstractModel {
	defaults() {
		return {
			name: '',
			emailRemindersEnabled: true,
			discoverableByName: false,
			discoverableByEmail: false,
			overdueTasksRemindersEnabled: true,
		}
	}
}