import AbstractService from '@/services/abstractService'
import NotificationModel from '@/models/notification'
import type {INotification} from '@/modelTypes/INotification'
import {toISOStringOrNull} from '@/helpers/time/toISOStringOrNull'

export default class NotificationService extends AbstractService<INotification> {
	constructor() {
		super({
			getAll: '/notifications',
			update: '/notifications/{id}',
			delete: '/notifications',
		})
	}

	modelFactory(data) {
		return new NotificationModel(data)
	}

	beforeUpdate(model) {
		if (!model) {
			return model
		}

		return {
			...model,
			created: toISOStringOrNull(model.created),
			readAt: toISOStringOrNull(model.readAt),
		}
	}
	
	async markAllRead() {
		return this.post('/notifications', false)
	}
}
