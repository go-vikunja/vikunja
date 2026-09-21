import AbstractModel from '@/models/abstractModel'
import type {IWebhook} from '@/modelTypes/IWebhook'

export default class WebhookModel extends AbstractModel<IWebhook> implements IWebhook {
	id = 0
	projectId = 0
	userId = 0
	secret = ''
	basicAuthUser = ''
	basicAuthPassword = ''
	targetUrl = ''
	events = []
	createdBy: IWebhook['createdBy'] = {}

	created: Date
	updated: Date

	constructor(data: Partial<IWebhook> = {}) {
		super()
		this.assignData(data)

		this.createdBy = data.createdBy ?? (data as {created_by?: IWebhook['createdBy']}).created_by ?? {}

		this.created = new Date(this.created)
		this.updated = new Date(this.updated)
	}
}
