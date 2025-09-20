import AbstractService from '@/services/abstractService'
import NotificationModel from '@/models/notification'
import type {INotification} from '@/modelTypes/INotification'

export default class NotificationService extends AbstractService<INotification> {
	constructor() {
		super({
			getAll: '/notifications',
			update: '/notifications/{id}',
		})
	}

	modelFactory(data: Partial<INotification>) {
		return new NotificationModel(data)
	}

	beforeUpdate(model: INotification) {
		if (!model) {
			return model
		}

		// Create a copy to avoid mutating the original model
		const processedModel = {...model}
		processedModel.created = new Date(model.created).toISOString() as any
		if (model.readAt) {
			processedModel.readAt = new Date(model.readAt).toISOString() as any
		}
		return processedModel
	}
	
	async markAllRead() {
		return this.post('/notifications', {} as INotification)
	}
}
