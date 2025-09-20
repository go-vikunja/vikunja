import AbstractModel from './abstractModel'
import UserSettingsModel from '@/models/userSettings'

import { AUTH_TYPES, type IUser, type AuthType } from '@/modelTypes/IUser'
import type { IUserSettings } from '@/modelTypes/IUserSettings'
import AvatarService from '@/services/avatar'

const avatarService = new AvatarService()
const avatarCache = new Map<string, string>()
const pendingRequests = new Map<string, Promise<string>>()

export async function fetchAvatarBlobUrl(user: IUser, size = 50) {
	if (!user || !user.username) {
		return ''
	}
	const key = `${user.username as string}-${size}`
	
	// Return cached URL if available
	if (avatarCache.has(key)) {
		return avatarCache.get(key) as string
	}
	
	// If there's already a pending request for this avatar, wait for it
	if (pendingRequests.has(key)) {
		return await (pendingRequests.get(key)!) as string
	}
	
	invalidateAvatarCache(user)
	
	// Create a new request
	const requestPromise = avatarService.getBlobUrl(`/avatar/${user.username}?size=${size}`)
		.then(url => {
			avatarCache.set(key, url)
			pendingRequests.delete(key)
			return url
		})
		.catch(error => {
			pendingRequests.delete(key)
			throw error
		})
	
	pendingRequests.set(key, requestPromise)
	return await requestPromise
}

export function invalidateAvatarCache(user: IUser) {
	if (!user || !user.username) {
		return
	}

	for (const key of Array.from(avatarCache.keys())) {
		if (key.startsWith(`${user.username}-`)) {
			avatarCache.delete(key)
		}
	}

	for (const key of Array.from(pendingRequests.keys())) {
		if (key.startsWith(`${user.username}-`)) {
			pendingRequests.delete(key)
		}
	}
}

export function getDisplayName(user: IUser) {
	if (user.name !== '') {
		return user.name
	}

	return user.username
}

export default class UserModel extends AbstractModel<IUser> implements IUser {
	// Index signature to make class compatible with Record<string, unknown>
	[key: string]: unknown

	id = 0
	email = ''
	username = ''
	name = ''
	exp = 0
	type: AuthType = AUTH_TYPES.UNKNOWN

	created: Date = new Date()
	updated: Date = new Date()
	settings: IUserSettings = new UserSettingsModel()

	isLocalUser: boolean = false
	deletionScheduledAt: string | null = null

	constructor(data: Partial<IUser> = {}) {
		super()
		this.assignData(data)

		this.created = this.created ? new Date(this.created) : new Date()
		this.updated = this.updated ? new Date(this.updated) : new Date()

		this.settings = new UserSettingsModel(this.settings || {})
	}
}
