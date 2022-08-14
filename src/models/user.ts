import AbstractModel, { type IAbstract } from './abstractModel'
import UserSettingsModel, { type IUserSettings } from '@/models/userSettings'

export interface IUser extends IAbstract {
	id: number
	email: string
	username: string
	name: string

	created: Date
	updated: Date
	settings: IUserSettings
}

export default class UserModel extends AbstractModel implements IUser {
	id = 0
	email = ''
	username = ''
	name = ''

	created: Date = null
	updated: Date = null
	settings: IUserSettings = null

	constructor(data: Partial<IUser>) {
		super()
		this.assignData(data)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)

		if (this.settings !== null) {
			this.settings = new UserSettingsModel(this.settings)
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