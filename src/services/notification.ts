import AbstractService from '@/services/abstractService'
import {formatISO} from 'date-fns'
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
		model.created = formatISO(new Date(model.created))
		model.readAt = formatISO(new Date(model.readAt))
		return model
	}
}