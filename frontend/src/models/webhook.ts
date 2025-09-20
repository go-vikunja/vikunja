import AbstractModel from '@/models/abstractModel'
import type {IWebhook} from '@/modelTypes/IWebhook'
import UserModel from '@/models/user'

export default class WebhookModel extends AbstractModel<IWebhook> implements IWebhook {
	id = 0
	projectId = 0
	secret = ''
	targetUrl = ''
	events = []
	createdBy = new UserModel()

	created: Date = new Date()
	updated: Date = new Date()

	constructor(data: Partial<IWebhook> = {}) {
		super()
		this.assignData(data)
		
		this.createdBy = this.createdBy ? new UserModel(this.createdBy) : new UserModel()

		this.created = this.created ? new Date(this.created) : new Date()
		this.updated = this.updated ? new Date(this.updated) : new Date()
	}
}
