import AbstractModel from './abstractModel'
import UserSettingsModel from '@/models/userSettings'

import { AUTH_TYPES, type IUser, type AuthType } from '@/modelTypes/IUser'
import type { IUserSettings } from '@/modelTypes/IUserSettings'
import AvatarService from '@/services/avatar'

const avatarService = new AvatarService()
const avatarCache = new Map<string, string>()

export async function fetchAvatarBlobUrl(user: IUser, size = 50) {
       const key = `${user.username}-${size}`
       if (avatarCache.has(key)) {
               return avatarCache.get(key) as string
       }
       const url = await avatarService.getBlobUrl(user, size)
       avatarCache.set(key, url)
       return url
}

export function getDisplayName(user: IUser) {
	if (user.name !== '') {
		return user.name
	}

	return user.username
}

export default class UserModel extends AbstractModel<IUser> implements IUser {
	id = 0
	email = ''
	username = ''
	name = ''
	exp = 0
	type: AuthType = AUTH_TYPES.UNKNOWN

	created: Date
	updated: Date
	settings: IUserSettings

	isLocalUser: boolean
	deletionScheduledAt: null

	constructor(data: Partial<IUser> = {}) {
		super()
		this.assignData(data)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)

		this.settings = new UserSettingsModel(this.settings || {})
	}
}
