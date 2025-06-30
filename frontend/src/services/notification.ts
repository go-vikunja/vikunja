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
		
		// Convert dates to ISO strings if they are not null
		if (model.created) {
			(model as any).created = new Date(model.created).toISOString()
		}
		if (model.readAt) {
			(model as any).readAt = new Date(model.readAt).toISOString()
		}
		return model
	}
	
	async markAllRead() {
		return this.post('/notifications', new NotificationModel({}))
	}
}
