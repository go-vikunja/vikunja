import type {IAbstract} from './IAbstract'

export interface IUserWebhookSetting extends IAbstract {
	id: number
	userId: number
	notificationType: string
	enabled: boolean
	targetUrl: string
	created: Date
	updated: Date
}

export interface IWebhookNotificationType {
	type: string
	description: string
}
