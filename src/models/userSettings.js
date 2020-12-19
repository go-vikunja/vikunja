
import AbstractModel from './abstractModel'

export default class UserSettingsModel extends AbstractModel {
	defaults() {
		return {
			name: '',
			emailRemindersEnabled: true,
		}
	}
}