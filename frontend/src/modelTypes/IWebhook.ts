import type {IAbstract} from './IAbstract'
import type {IUser} from '@/modelTypes/IUser'

export interface IWebhook extends IAbstract {
	id: number
	projectId: number
  secret: string
	basicauthuser: string
  basicauthpassword: string
	targetUrl: string
	events: string[]
	createdBy: IUser

	created: Date
	updated: Date
}
