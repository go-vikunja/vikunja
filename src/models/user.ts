import AbstractModel from './abstractModel'
import UserSettingsModel from '@/models/userSettings'

import { AUTH_TYPES, type IUser } from '@/modelTypes/IUser'
import type { IUserSettings } from '@/modelTypes/IUserSettings'

export default class UserModel extends AbstractModel<IUser> implements IUser {
	id = 0
	email = ''
	username = ''
	name = ''
	exp = 0
	type = AUTH_TYPES.UNKNOWN

	created: Date
	updated: Date
	settings: IUserSettings

	constructor(data: Partial<IUser> = {}) {
		super()
		this.assignData(data)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)

		this.settings = new UserSettingsModel(this.settings || {})
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