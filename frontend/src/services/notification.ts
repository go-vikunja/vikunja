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

	modelFactory(data) {
		return new NotificationModel(data)
	}

	beforeUpdate(model) {
		if (!model) {
			return model
		}
		
		model.created = new Date(model.created).toISOString()
		model.readAt = new Date(model.readAt).toISOString()
		return model
	}
	
	async markAllRead() {
		return this.post('/notifications', false)
	}
}
