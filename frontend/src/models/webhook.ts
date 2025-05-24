import AbstractModel from '@/models/abstractModel'
import type {IWebhook} from '@/modelTypes/IWebhook'
import UserModel from '@/models/user'

export default class WebhookModel extends AbstractModel<IWebhook> implements IWebhook {
	id = 0
	projectId = 0
	secret = ''
	targetUrl = ''
	events = []
	createdBy = null

	created: Date
	updated: Date

	constructor(data: Partial<IWebhook> = {}) {
		super()
		this.assignData(data)
		
		this.createdBy = new UserModel(this.createdBy)

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
