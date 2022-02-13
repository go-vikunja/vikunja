import AbstractModel from './abstractModel'
import UserSettingsModel from '@/models/userSettings'

export default class UserModel extends AbstractModel {
	constructor(data) {
		super(data)

		/** @type {number} */
		this.id

		/** @type {string} */
		this.email

		/** @type {string} */
		this.username

		/** @type {string} */
		this.name

		/** @type {Date} */
		this.created = new Date(this.created)

		/** @type {Date} */
		this.updated = new Date(this.updated)

		/** @type {UserSettingsModel} */
		this.settings

		if (this.settings !== null) {
			this.settings = new UserSettingsModel(this.settings)
		}
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