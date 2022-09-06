import type {IAbstract} from './IAbstract'
import type {IUserSettings} from './IUserSettings'

export interface IUser extends IAbstract {
	id: number
	email: string
	username: string
	name: string

	created: Date
	updated: Date
	settings: IUserSettings
}