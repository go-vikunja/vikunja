import type {IAbstract} from './IAbstract'
import type {IUser} from '@/modelTypes/IUser'

export interface IWebhook extends IAbstract {
	id: number
	projectId: number
	secret: string
	targetUrl: string
	events: string[]
	createdBy: IUser

	created: Date
	updated: Date
}
