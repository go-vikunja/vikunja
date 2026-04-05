import type {IAbstract} from './IAbstract'

export interface IBotUser extends IAbstract {
	id: number
	username: string
	name: string
	status: number
	botOwnerId: number
	created: Date
	updated: Date
}
