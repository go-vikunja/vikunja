import AbstractModel from './abstractModel'
import UserSettingsModel from '@/models/userSettings'

export default class UserModel extends AbstractModel {
	constructor(data) {
		super(data)

		if (this.settings !== null) {
			this.settings = new UserSettingsModel(this.settings)
		}

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}

	defaults() {
		return {
			id: 0,
			email: '',
			username: '',
			name: '',
			created: null,
			updated: null,
			settings: null,
		}
	}

	getAvatarUrl(size = 50) {
		return `${window.API_URL}/avatar/${this.username}?size=${size}`
	}

	getDisplayName() {
		if (this.name !== '') {
			return this.name
		}

		return this.username
	}
}