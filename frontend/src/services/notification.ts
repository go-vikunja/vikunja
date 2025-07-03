import AbstractService from '@/services/abstractService'
import NotificationModel from '@/models/notification'
import type {INotification} from '@/modelTypes/INotification'

export default class NotificationService extends AbstractService<NotificationModel> {
	constructor() {
		super({
			getAll: '/notifications',
			update: '/notifications/{id}',
		})
	}

	modelFactory(data: Partial<INotification>) {
		return new NotificationModel(data)
	}

	modelGetAllFactory(data: Partial<INotification>) {
		return new NotificationModel(data)
	}

	beforeUpdate(model: NotificationModel) {
		if (!model) {
			return model
		}
		
		// Create a serializable copy with ISO date strings for API transmission
		const serializedModel = {
			...model,
			created: model.created ? new Date(model.created).toISOString() : model.created,
			readAt: model.readAt ? new Date(model.readAt).toISOString() : model.readAt,
		}
		
		// Return as NotificationModel type since the API expects this format
		return serializedModel as unknown as NotificationModel
	}
	
	async markAllRead() {
		return this.post('/notifications', new NotificationModel({}))
	}
}
