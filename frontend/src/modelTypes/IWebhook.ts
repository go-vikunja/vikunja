import type {IAbstract} from './IAbstract'
import type {IUser} from '@/modelTypes/IUser'

export interface IWebhook extends IAbstract {
	id: number
	projectId: number
	userId: number
	secret: string
	basicAuthUser: string
	basicAuthPassword: string
	targetUrl: string
	events: string[]
	createdBy: IUser

	created: Date
	updated: Date
}
