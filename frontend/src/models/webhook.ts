import AbstractModel from '@/models/abstractModel'
import type {IWebhook} from '@/modelTypes/IWebhook'
import UserModel from '@/models/user'

export default class WebhookModel extends AbstractModel<IWebhook> implements IWebhook {
	id = 0
	projectId = 0
	secret = ''
	targetUrl = ''
	events = []
	createdBy: any = null

	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<IWebhook> = {}) {
		super()
		this.assignData(data)
		
		if (this.createdBy) {
			this.createdBy = new UserModel(this.createdBy)
		}

		this.created = new Date(this.created || Date.now())
		this.updated = new Date(this.updated || Date.now())
	}
}
