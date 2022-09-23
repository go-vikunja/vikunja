import type {IAbstract} from './IAbstract'
import type {IUserSettings} from './IUserSettings'

export const AUTH_TYPES = {
	'UNKNOWN': 0,
	'USER': 1,
	'LINK_SHARE': 2,
} as const

export interface IUser extends IAbstract {
	id: number
	email: string
	username: string
	name: string
	exp: number
	type: typeof AUTH_TYPES[keyof typeof AUTH_TYPES],

	created: Date
	updated: Date
	settings: IUserSettings
}