import AbstractModel from './abstractModel'
import UserSettingsModel, { type IUserSettings } from '@/models/userSettings'

export interface IUser {
	id: number
	email: string
	username: string
	name: string

	created: Date
	updated: Date
	settings: IUserSettings
}

export default class UserModel extends AbstractModel implements IUser {
	declare id: number
	declare email: string
	declare username: string
	declare name: string

	created: Date
	updated: Date
	settings: IUserSettings

	constructor(data) {
		super(data)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)

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