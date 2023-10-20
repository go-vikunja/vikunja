import AbstractService from '@/services/abstractService'
import type {IWebhook} from '@/modelTypes/IWebhook'
import WebhookModel from '@/models/webhook'

export default class WebhookService extends AbstractService<IWebhook> {
	constructor() {
		super({
			getAll: '/projects/{projectId}/webhooks',
			create: '/projects/{projectId}/webhooks',
			update: '/projects/{projectId}/webhooks/{id}',
			delete: '/projects/{projectId}/webhooks/{id}',
		})
	}

	modelFactory(data) {
		return new WebhookModel(data)
	}

	async getAvailableEvents(): Promise<string[]> {
		const cancel = this.setLoading()

		try {
			const response = await this.http.get('/webhooks/events')
			return response.data
		} finally {
			cancel()
		}
	}
}
